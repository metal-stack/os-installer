package v1

import (
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
)

// Bootinfo is written by the installer in the target os to tell us
// which kernel, initrd and cmdline must be used for kexec
type Bootinfo struct {
	Initrd       string `yaml:"initrd"`
	Cmdline      string `yaml:"cmdline"`
	Kernel       string `yaml:"kernel"`
	BootloaderID string `yaml:"bootloader_id"`
}

const InstallerConfigPath = "/etc/metal/os-installer.yaml"

// InstallerConfig can be placed inside the target OS to customize the os-installer.
type InstallerConfig struct {
	// OsName enforces a specific os-installer implementation, defaults to auto-detection
	OsName *string `yaml:"os_name"`
	// Only allows to run installer tasks only with the given names
	Only []string `yaml:"only"`
	// Except allows to run installer tasks except for the given names
	Except []string `yaml:"except"`
	// CustomScript allows executing a custom script that's placed inside the OS at the end of the installer execution
	CustomScript *struct {
		ExecutablePath string `yaml:"executable_path"`
		WorkDir        string `yaml:"workdir"`
	} `yaml:"custom_script"`
	// Overwrites allows specifying os-installer overwrites for the default implementation
	Overwrites struct {
		BootloaderID *string `yaml:"bootloader_id"`
	}
}

type MachineDetails struct {
	// Id is the machine UUID
	ID string `yaml:"id"`
	// Nics are the nics of the machine
	Nics []*apiv2.MachineNic `yaml:"nics"`
	// Password is the password for the metal user.
	Password string `yaml:"password"`
	// Console specifies where the kernel should connect its console to.
	Console string `yaml:"console"`
	// RaidEnabled is set to true if any raid devices are specified
	RaidEnabled bool `yaml:"raidenabled"`
	// RootUUID is the fs uuid if the root fs
	RootUUID string `yaml:"root_uuid"`
}

// FIXME legacy structs remove once old images are gone

type (
	// Disk is a physical Disk
	Disk struct {
		// Device the name of the disk device visible from kernel side, e.g. sda
		Device string
		// Partitions to create on this disk, order is preserved
		Partitions []Partition
	}
	Partition struct {
		Label      string
		Filesystem string
		Properties map[string]string
	}
)
