package oscommon

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/metal-stack/os-installer/pkg/exec"
	"github.com/spf13/afero"
)

const (
	DefaultGrubPath        = "/etc/default/grub"
	defaultGrubFileContent = `GRUB_DEFAULT=0
GRUB_TIMEOUT=5
GRUB_DISTRIBUTOR=%s
GRUB_CMDLINE_LINUX_DEFAULT=""
GRUB_CMDLINE_LINUX="%s"
GRUB_TERMINAL=serial
GRUB_SERIAL_COMMAND="serial --speed=%s --unit=%s --word=8"
`
)

func (d *CommonTasks) GrubInstall(ctx context.Context, bootloaderID, cmdLine string) error {
	serialPort, serialSpeed, err := d.FigureOutSerialSpeed()
	if err != nil {
		return err
	}

	defaultGrub := fmt.Sprintf(defaultGrubFileContent, bootloaderID, cmdLine, serialSpeed, serialPort)

	err = d.fs.WriteFile(DefaultGrubPath, []byte(defaultGrub), 0755)
	if err != nil {
		return err
	}

	grubInstallArgs := []string{
		"--target=x86_64-efi",
		"--efi-directory=/boot/efi",
		"--boot-directory=/boot",
		"--bootloader-id=" + bootloaderID,
		"--removable",
	}

	if d.details.RaidEnabled {
		grubInstallArgs = append(grubInstallArgs, "--no-nvram")

		out, err := d.exec.Execute(ctx, &exec.Params{
			Name:    "mdadm",
			Args:    []string{"--examine", "--scan"},
			Timeout: 10 * time.Second,
		})
		if err != nil {
			return err
		}

		out += "\nMAILADDR root\n"

		err = afero.WriteFile(d.fs, "/etc/mdadm.conf", []byte(out), 0700)
		if err != nil {
			return err
		}

		err = d.fs.MkdirAll("/var/lib/initramfs-tools", 0755)
		if err != nil {
			return err
		}

		_, err = d.exec.Execute(ctx, &exec.Params{
			Name: "update-initramfs",
			Args: []string{"-u"},
		})
		if err != nil {
			return err
		}

		out, err = d.exec.Execute(ctx, &exec.Params{
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

				shim := fmt.Sprintf(`\\EFI\\%s\\grubx64.efi`, bootloaderID)

				_, err = d.exec.Execute(ctx, &exec.Params{
					Name: "efibootmgr",
					Args: []string{"-c", "-d", disk, "-p1", "-l", shim, "-L", bootloaderID},
				})
				if err != nil {
					return err
				}
			}
		}
	}

	if !runFromCI() {
		_, err = d.exec.Execute(ctx, &exec.Params{
			Name: "grub-install",
			Args: grubInstallArgs,
		})
		if err != nil {
			return err
		}
	}

	_, err = d.exec.Execute(ctx, &exec.Params{
		Name: "update-grub2",
	})
	if err != nil {
		return err
	}

	_, err = d.exec.Execute(ctx, &exec.Params{
		Name: "dpkg-reconfigure",
		Args: []string{"grub-efi-amd64-bin"},
		Env: []string{
			"DEBCONF_NONINTERACTIVE_SEEN=true",
			"DEBIAN_FRONTEND=noninteractive",
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (d *CommonTasks) FigureOutSerialSpeed() (serialPort, serialSpeed string, err error) {
	// ttyS1,115200n8
	serialPort, serialSpeed, found := strings.Cut(d.details.Console, ",")
	if !found {
		return "", "", fmt.Errorf("serial console could not be split into port and speed")
	}

	_, serialPort, found = strings.Cut(serialPort, "ttyS")
	if !found {
		return "", "", fmt.Errorf("serial port could not be split")
	}

	serialSpeed, _, found = strings.Cut(serialSpeed, "n8")
	if !found {
		return "", "", fmt.Errorf("serial speed could not be split")
	}

	return
}
