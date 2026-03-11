package oscommon

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strconv"
	"strings"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	v1 "github.com/metal-stack/os-installer/api/v1"
	"github.com/metal-stack/os-installer/pkg/exec"
	"github.com/metal-stack/os-installer/pkg/network"
	"github.com/spf13/afero"
)

type (
	OperatingSystem interface {
		WriteHostname(ctx context.Context) error
		WriteHosts(ctx context.Context) error
		WriteResolvConf(ctx context.Context) error
		WriteNTPConf(ctx context.Context) error
		CreateMetalUser(ctx context.Context) error
		ConfigureNetwork(ctx context.Context) error
		CopySSHKeys(ctx context.Context) error
		FixPermissions(ctx context.Context) error
		ProcessUserdata(ctx context.Context) error
		BuildCMDLine(ctx context.Context) (string, error)
		WriteBootInfo(ctx context.Context, cmdLine string) error
		GrubInstall(ctx context.Context, cmdLine string) error
		UnsetMachineID(ctx context.Context) error
		SystemdServices(ctx context.Context) error
		WriteBuildMeta(ctx context.Context) error

		SudoGroup() string
		InitramdiskFormatString() string
		BootloaderID() string
	}

	Config struct {
		Log            *slog.Logger
		Fs             *afero.Afero
		Exec           *exec.CmdExecutor
		MachineDetails *v1.MachineDetails
		Allocation     *apiv2.MachineAllocation
	}

	DefaultOS struct {
		log        *slog.Logger
		fs         *afero.Afero
		details    *v1.MachineDetails
		allocation *apiv2.MachineAllocation
		exec       *exec.CmdExecutor
		network    *network.Network
	}
)

func New(cfg *Config) *DefaultOS {
	return &DefaultOS{
		log:        cfg.Log,
		fs:         cfg.Fs,
		details:    cfg.MachineDetails,
		allocation: cfg.Allocation,
		exec:       cfg.Exec,
		network:    network.New(cfg.Allocation),
	}
}

func (d *DefaultOS) SudoGroup() string {
	return "sudo"
}

func (d *DefaultOS) BootloaderID() string {
	panic("default os does not provide a bootloader id")
}

func (d *DefaultOS) InitramdiskFormatString() string {
	return "initrd.img-%s"
}

func (d *DefaultOS) GetKernelVersion(initramdiskFormatString string) (string, error) {
	kern, _, err := d.KernelAndInitrdPath(initramdiskFormatString)
	if err != nil {
		return "", err
	}

	_, version, found := strings.Cut(kern, "vmlinuz-")
	if !found {
		return "", fmt.Errorf("unable to determine kernel version from: %s", kern)
	}

	return version, nil
}

func (d *DefaultOS) KernelAndInitrdPath(initramdiskFormatString string) (kern string, initrd string, err error) {
	// Debian 10
	// root@1f223b59051bcb12:/boot# ls -l
	// total 83500
	// -rw-r--r-- 1 root root       83 Aug 13 15:25 System.map-5.10.0-17-amd64
	// -rw-r--r-- 1 root root   236286 Aug 13 15:25 config-5.10.0-17-amd64
	// -rw-r--r-- 1 root root    93842 Jul 19  2021 config-5.10.51
	// drwxr-xr-x 2 root root     4096 Oct  3 11:21 grub
	// -rw-r--r-- 1 root root 34665690 Oct  3 11:22 initrd.img-5.10.0-17-amd64
	// lrwxrwxrwx 1 root root       21 Jul 19  2021 vmlinux -> /boot/vmlinux-5.10.51
	// -rwxr-xr-x 1 root root 43526368 Jul 19  2021 vmlinux-5.10.51
	// -rw-r--r-- 1 root root  6962816 Aug 13 15:25 vmlinuz-5.10.0-17-amd64

	// Ubuntu 20.04
	// root@568551f94559b121:~# ls -l /boot/
	// total 83500
	// -rw-r--r-- 1 root root       83 Aug 13 15:25 System.map-5.10.0-17-amd64
	// -rw-r--r-- 1 root root   236286 Aug 13 15:25 config-5.10.0-17-amd64
	// -rw-r--r-- 1 root root    93842 Jul 19  2021 config-5.10.51
	// drwxr-xr-x 2 root root     4096 Oct  3 11:21 grub
	// -rw-r--r-- 1 root root 34665690 Oct  3 11:22 initrd.img-5.10.0-17-amd64
	// lrwxrwxrwx 1 root root       21 Jul 19  2021 vmlinux -> /boot/vmlinux-5.10.51
	// -rwxr-xr-x 1 root root 43526368 Jul 19  2021 vmlinux-5.10.51
	// -rw-r--r-- 1 root root  6962816 Aug 13 15:25 vmlinuz-5.10.0-17-amd64

	// Almalinux 9
	// [root@14231d4e67d28390 ~]# ls -l /boot/
	// total 160420
	// -rw------- 1 root root  8876661 Jan  7 23:19 System.map-5.14.0-503.19.1.el9_5.x86_64
	// -rw-r--r-- 1 root root    93842 Jul 19  2021 config-5.10.51
	// -rw-r--r-- 1 root root   226249 Jan  7 23:19 config-5.14.0-503.19.1.el9_5.x86_64
	// drwx------ 3 root root     4096 Jun  8  2022 efi
	// drwx------ 3 root root     4096 Jan  9 08:02 grub2
	// -rw------- 1 root root 97054329 Jan  9 08:04 initramfs-5.14.0-503.19.1.el9_5.x86_64.img
	// drwxr-xr-x 3 root root     4096 Jan  9 08:02 loader
	// lrwxrwxrwx 1 root root       52 Jan  9 08:03 symvers-5.14.0-503.19.1.el9_5.x86_64.gz -> /lib/modules/5.14.0-503.19.1.el9_5.x86_64/symvers.gz
	// lrwxrwxrwx 1 root root       21 Jul 19  2021 vmlinux -> /boot/vmlinux-5.10.51
	// -rwxr-xr-x 1 root root 43526368 Jul 19  2021 vmlinux-5.10.51
	// -rwxr-xr-x 1 root root 14467384 Jan  7 23:19 vmlinuz-5.14.0-503.19.1.el9_5.x86_64

	var (
		bootPartition   = "/boot"
		systemMapPrefix = "/boot/System.map-"
	)

	systemMaps, err := afero.Glob(d.fs, systemMapPrefix+"*")
	if err != nil {
		return "", "", fmt.Errorf("unable to find a System.map, probably no kernel installed: %w", err)
	}
	if len(systemMaps) != 1 {
		return "", "", fmt.Errorf("more or less than a single System.map found (%v), probably no kernel or more than one kernel installed", systemMaps)
	}

	systemMap := systemMaps[0]
	_, kernelVersion, found := strings.Cut(systemMap, systemMapPrefix)
	if !found {
		return "", "", fmt.Errorf("unable to detect kernel version in System.map: %q", systemMap)
	}

	kern = path.Join(bootPartition, "vmlinuz"+"-"+kernelVersion)
	if !d.fileExists(kern) {
		return "", "", fmt.Errorf("kernel image %q not found", kern)
	}

	initrd = path.Join(bootPartition, fmt.Sprintf(initramdiskFormatString, kernelVersion))
	if !d.fileExists(initrd) {
		return "", "", fmt.Errorf("ramdisk %q not found", initrd)
	}

	d.log.Info("detect kernel and initrd", "kernel", kern, "initrd", initrd)

	return
}

func (d *DefaultOS) fileExists(filename string) bool {
	info, err := d.fs.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func runFromCI() bool {
	ciEnv := os.Getenv("INSTALL_FROM_CI")

	ci, err := strconv.ParseBool(ciEnv)
	if err != nil {
		return false
	}

	return ci
}
