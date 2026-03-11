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
