package os

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	v1 "github.com/metal-stack/os-installer/api/v1"
	operatingsystem "github.com/metal-stack/os-installer/pkg/installer/os"
	oscommon "github.com/metal-stack/os-installer/pkg/installer/os/common"
	"github.com/spf13/afero"
)

type installer struct {
	log *slog.Logger
	oss oscommon.OperatingSystem
	fs  *afero.Afero
}

func Install(ctx context.Context, log *slog.Logger, details *v1.MachineDetails, allocation *apiv2.MachineAllocation) error {
	log = log.WithGroup("os-installer")

	var (
		start = time.Now()
		fs    = &afero.Afero{
			Fs: afero.OsFs{},
		}
	)

	oss, err := operatingsystem.New(&oscommon.Config{
		Log:            log,
		Fs:             fs,
		MachineDetails: details,
		Allocation:     allocation,
	})
	if err != nil {
		return fmt.Errorf("os detection failed: %w", err)
	}

	i := installer{
		log: log,
		oss: oss,
		fs:  fs,
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
	} {
		var (
			log   = i.log.With("task-name", task.name)
			start = time.Now()
		)

		log.Info("running install task", "start-at", start.String())

		if err := task.fn(ctx); err != nil {
			i.log.Info("running install task failed", "took", time.Since(start).String())
			return fmt.Errorf("installation task failed, aborting install: %w", err)
		}
	}

	return nil
}

func (i *installer) validateRunningInEfiMode(ctx context.Context) error {
	if !i.isVirtual() && !i.fileExists("/sys/firmware/efi") {
		return fmt.Errorf("not running efi mode")
	}

	return nil
}

func (i *installer) removeDockerEnv(_ context.Context) error {
	// systemd-detect-virt guesses docker which modifies the behavior of many services.
	if !i.fileExists("/.dockerenv") {
		return nil
	}

	return i.fs.Remove("/.dockerenv")
}

func (i *installer) isVirtual() bool {
	return !i.fileExists("/sys/class/dmi")
}

func (i *installer) fileExists(filename string) bool {
	info, err := i.fs.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}
