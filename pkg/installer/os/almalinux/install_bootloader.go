package almalinux

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/metal-stack/os-installer/pkg/exec"
)

const (
	defaultGrubPath        = "/etc/default/grub"
	defaultGrubFileContent = `GRUB_DEFAULT=0
GRUB_TIMEOUT=5
GRUB_DISTRIBUTOR=%s
GRUB_CMDLINE_LINUX_DEFAULT=""
GRUB_CMDLINE_LINUX="%s"
GRUB_TERMINAL=serial
GRUB_SERIAL_COMMAND="serial --speed=%s --unit=%s --word=8"
GRUB_DEVICE=UUID=%s
GRUB_ENABLE_BLSCFG=false
`
	grubConfigPath = "/boot/efi/EFI/almalinux/grub.cfg"
)

func (o *os) GrubInstall(ctx context.Context, cmdLine string) error {
	serialSpeed, serialPort, err := o.FigureOutSerialSpeed()
	if err != nil {
		return err
	}

	defaultGrub := fmt.Sprintf(defaultGrubFileContent, o.BootloaderID(), cmdLine, serialSpeed, serialPort, o.details.RootUUID)

	err = o.fs.WriteFile(defaultGrubPath, []byte(defaultGrub), 0755)
	if err != nil {
		return err
	}

	grubInstallArgs := []string{
		"--target=x86_64-efi",
		"--efi-directory=/boot/efi",
		"--boot-directory=/boot",
		"--bootloader-id=" + o.BootloaderID(),
	}
	if o.details.RaidEnabled {
		grubInstallArgs = append(grubInstallArgs, "--no-nvram")
	}

	_, err = o.exec.Execute(ctx, &exec.Params{
		Name: "grub2-mkconfig",
		Args: []string{"-o", grubConfigPath},
	})
	if err != nil {
		return err
	}

	grubInstallArgs = append(grubInstallArgs, fmt.Sprintf("UUID=%s", o.details.RootUUID))

	if o.details.RaidEnabled {
		out, err := o.exec.Execute(ctx, &exec.Params{
			Name:    "mdadm",
			Args:    []string{"--examine", "--scan"},
			Timeout: 10 * time.Second,
		})
		if err != nil {
			return err
		}

		out += "\nMAILADDR root\n"

		err = o.fs.WriteFile("/etc/mdadm.conf", []byte(out), 0700)
		if err != nil {
			return err
		}

		out, err = o.exec.Execute(ctx, &exec.Params{
			Name: "blkid",
		})
		if err != nil {
			return err
		}

		for line := range strings.SplitSeq(string(out), "\n") {
			if strings.Contains(line, `PARTLABEL="efi"`) {
				disk, _, found := strings.Cut(line, ":")
				if !found {
					return fmt.Errorf("unable to process blkid output lines")
				}

				shim := fmt.Sprintf(`\\EFI\\%s\\shimx64.efi`, o.BootloaderID())

				_, err = o.exec.Execute(ctx, &exec.Params{
					Name: "efibootmgr",
					Args: []string{"-c", "-d", disk, "-p1", "-l", shim, "-L", o.BootloaderID()},
				})
				if err != nil {
					return err
				}
			}
		}
	}

	if !o.details.RaidEnabled {
		return nil
	}

	v, err := o.GetKernelVersion(o.InitramdiskFormatString())
	if err != nil {
		return err
	}

	_, err = o.exec.Execute(ctx, &exec.Params{
		Name: "dracut",
		Args: []string{
			"--mdadmconf",
			"--kver", v,
			"--kmoddir", "/lib/modules/" + v,
			"--include", "/lib/modules/" + v, "/lib/modules/" + v,
			"--fstab",
			"--add=dm mdraid",
			"--add-drivers=raid0 raid1",
			"--hostonly",
			"--force",
		},
	})
	if err != nil {
		return err
	}

	return nil
}
