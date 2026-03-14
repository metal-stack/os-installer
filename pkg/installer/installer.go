package installer

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"time"

	"buf.build/go/protoyaml"
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	v1 "github.com/metal-stack/os-installer/api/v1"
	"github.com/metal-stack/os-installer/pkg/exec"
	operatingsystem "github.com/metal-stack/os-installer/pkg/installer/os"
	oscommon "github.com/metal-stack/os-installer/pkg/installer/os/common"
	"github.com/spf13/afero"
	"go.yaml.in/yaml/v3"
)

type installer struct {
	log        *slog.Logger
	cfg        *v1.Config
	oss        oscommon.OperatingSystem
	fs         *afero.Afero
	exec       *exec.CmdExecutor
	details    *v1.MachineDetails
	allocation *apiv2.MachineAllocation
}

func New(log *slog.Logger, details *v1.MachineDetails, allocation *apiv2.MachineAllocation) *installer {
	return &installer{
		log: log.WithGroup("os-installer"),
		cfg: &v1.Config{},
		fs: &afero.Afero{
			Fs: afero.OsFs{},
		},
		details:    details,
		allocation: allocation,
	}
}

func (i *installer) PersistConfigurations() error {
	detailsBytes, err := yaml.Marshal(i.details)
	if err != nil {
		return fmt.Errorf("unable to marshal machine details: %w", err)
	}
	err = os.WriteFile(v1.MachineDetailsPath, detailsBytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to persist machine details: %w", err)
	}

	allocationBytes, err := protoyaml.Marshal(i.allocation)
	if err != nil {
		return fmt.Errorf("unable to marshal machine allocation: %w", err)
	}
	err = os.WriteFile(v1.MachineAllocationPath, allocationBytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to persist machine allocation: %w", err)
	}
	return nil
}

func ReadConfigurations() (*v1.MachineDetails, *apiv2.MachineAllocation, error) {
	data, err := os.ReadFile(v1.MachineDetailsPath)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read machine details: %w", err)
	}

	var details v1.MachineDetails
	if err = yaml.Unmarshal(data, &details); err != nil {
		return nil, nil, fmt.Errorf("unable to parse machine details: %w", err)
	}

	data, err = os.ReadFile(v1.MachineAllocationPath)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read machine allocation: %w", err)
	}

	var allocation apiv2.MachineAllocation
	if err = protoyaml.Unmarshal(data, &allocation); err != nil {
		return nil, nil, fmt.Errorf("unable to parse machine allocation: %w", err)
	}

	return &details, &allocation, nil
}

func (i *installer) Install(ctx context.Context) error {
	var (
		start           = time.Now()
		installerConfig = &v1.Config{}
	)

	if oscommon.FileExists(i.fs, v1.InstallerConfigPath) {
		data, err := i.fs.ReadFile(v1.InstallerConfigPath)
		if err != nil {
			return fmt.Errorf("unable to read installer config: %w", err)
		}

		if err = yaml.Unmarshal(data, &installerConfig); err != nil {
			return fmt.Errorf("unable to parse installer config: %w", err)
		}
	}

	oss, err := operatingsystem.New(&oscommon.Config{
		Log:            i.log,
		Fs:             i.fs,
		MachineDetails: i.details,
		Allocation:     i.allocation,
		Name:           installerConfig.OsName,
		BootloaderID:   installerConfig.Overwrites.BootloaderID,
	})
	if err != nil {
		return fmt.Errorf("os detection failed: %w", err)
	}

	i.cfg = installerConfig
	i.oss = oss
	i.exec = exec.New(i.log)

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
