package os

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	v1 "github.com/metal-stack/os-installer/api/v1"
	"github.com/metal-stack/os-installer/pkg/exec"
	operatingsystem "github.com/metal-stack/os-installer/pkg/installer/os"
	oscommon "github.com/metal-stack/os-installer/pkg/installer/os/common"
	"github.com/spf13/afero"
	"go.yaml.in/yaml/v3"
)

type installer struct {
	log  *slog.Logger
	cfg  *v1.InstallerConfig
	oss  oscommon.OperatingSystem
	fs   *afero.Afero
	exec *exec.CmdExecutor
}

func Install(ctx context.Context, log *slog.Logger, details *v1.MachineDetails, allocation *apiv2.MachineAllocation) error {
	log = log.WithGroup("os-installer")

	var (
		start = time.Now()
		fs    = &afero.Afero{
			Fs: afero.OsFs{},
		}
		installerConfig = &v1.InstallerConfig{}
	)

	if oscommon.FileExists(fs, v1.InstallerConfigPath) {
		data, err := fs.ReadFile(v1.InstallerConfigPath)
		if err != nil {
			return fmt.Errorf("unable to read installer config: %w", err)
		}

		if err = yaml.Unmarshal(data, &installerConfig); err != nil {
			return fmt.Errorf("unable to parse installer config: %w", err)
		}
	}

	oss, err := operatingsystem.New(&oscommon.Config{
		Log:            log,
		Fs:             fs,
		MachineDetails: details,
		Allocation:     allocation,
		Name:           installerConfig.OsName,
		BootloaderID:   installerConfig.Overwrites.BootloaderID,
	})
	if err != nil {
		return fmt.Errorf("os detection failed: %w", err)
	}

	i := installer{
		log:  log,
		cfg:  installerConfig,
		oss:  oss,
		exec: exec.New(log),
		fs:   fs,
	}

	if err = i.run(ctx); err != nil {
		i.log.Info("running os installer failed", "took", time.Since(start).String())
		return fmt.Errorf("os installer failed: %w", err)
	}

	i.log.Info("os installer succeeded", "took", time.Since(start).String())

	return nil
}

func (i *installer) run(ctx context.Context) error {
	var (
		cmdLine string
	)

	for _, task := range []struct {
		name string
		fn   func(ctx context.Context) error
	}{
		{
			name: "check if running in efi mode",
			fn:   i.validateRunningInEfiMode,
		},
		{
			name: "remove .dockerenv if running in virtual environment",
			fn:   i.removeDockerEnv,
		},
		{
			name: "write hostname",
			fn:   i.oss.WriteHostname,
		},
		{
			name: "write /etc/hosts",
			fn:   i.oss.WriteHosts,
		},
		{
			name: "write /etc/resolv.conf",
			fn:   i.oss.WriteResolvConf,
		},
		{
			name: "write ntp configuration",
			fn:   i.oss.WriteNTPConf,
		},
		{
			name: "create metal user",
			fn:   i.oss.CreateMetalUser,
		},
		{
			name: "configure network",
			fn:   i.oss.ConfigureNetwork,
		},
		{
			name: "authorized ssh keys",
			fn:   i.oss.CopySSHKeys,
		},
		{
			name: "fix wrong filesystem permissions",
			fn:   i.oss.FixPermissions,
		},
		{
			name: "build kernel cmdline",
			fn: func(ctx context.Context) error {
				l, err := i.oss.BuildCMDLine(ctx)
				if err != nil {
					return err
				}

				cmdLine = l

				return nil
			},
		},
		{
			name: "write /etc/metal/boot-info.yaml",
			fn: func(ctx context.Context) error {
				return i.oss.WriteBootInfo(ctx, cmdLine)
			},
		},
		{
			name: "write booatloader config",
			fn: func(ctx context.Context) error {
				return i.oss.GrubInstall(ctx, cmdLine)
			},
		},
		{
			name: "unset machine id",
			fn:   i.oss.UnsetMachineID,
		},
		{
			name: "deploy systemd services",
			fn:   i.oss.SystemdServices,
		},
		{
			name: "write /etc/metal/build-meta.yaml",
			fn:   i.oss.WriteBuildMeta,
		},
		{
			name: "execute custom executable",
			fn:   i.customExecutable,
		},
	} {
		var (
			log   = i.log.With("task-name", task.name)
			start = time.Now()
		)

		if len(i.cfg.Only) > 0 && !slices.Contains(i.cfg.Only, task.name) {
			log.Info("skipping task as defined by installer configuration")
			continue
		}

		if slices.Contains(i.cfg.Except, task.name) {
			log.Info("skipping task as defined by installer configuration")
			continue
		}

		log.Info("running install task", "start-at", start.String())

		if err := task.fn(ctx); err != nil {
			i.log.Info("running install task failed", "took", time.Since(start).String())
			return fmt.Errorf("installation task failed, aborting install: %w", err)
		}
	}

	return nil
}

func (i *installer) validateRunningInEfiMode(ctx context.Context) error {
	if !i.isVirtual() && !oscommon.FileExists(i.fs, "/sys/firmware/efi") {
		return fmt.Errorf("not running efi mode")
	}

	return nil
}

func (i *installer) removeDockerEnv(_ context.Context) error {
	// systemd-detect-virt guesses docker which modifies the behavior of many services.
	if !oscommon.FileExists(i.fs, "/.dockerenv") {
		return nil
	}

	return i.fs.Remove("/.dockerenv")
}

func (i *installer) isVirtual() bool {
	return !oscommon.FileExists(i.fs, "/sys/class/dmi")
}

func (i *installer) customExecutable(ctx context.Context) error {
	if i.cfg.CustomScript == nil {
		i.log.Info("no custom executable to execute, skipping")
		return nil
	}

	_, err := i.exec.Execute(ctx, &exec.Params{
		Name: i.cfg.CustomScript.ExecutablePath,
		Dir:  i.cfg.CustomScript.WorkDir,
	})
	if err != nil {
		return fmt.Errorf("custom executable returned an error code: %w", err)
	}

	return nil
}
