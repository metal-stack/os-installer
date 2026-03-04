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

// InstallerConfig contains configuration items which are
// used to install the os.
type InstallerConfig struct {
	// Hostname of the machine
	Hostname string `yaml:"hostname"`
	// Networks all networks connected to this machine
	Networks []*apiv2.MachineNetwork `yaml:"networks"`
	// MachineUUID is the unique UUID for this machine, usually the board serial.
	MachineUUID string `yaml:"machineuuid"`
	// SSHPublicKey of the user
	SSHPublicKey string `yaml:"sshpublickey"`
	// Password is the password for the metal user.
	Password string `yaml:"password"`
	// Console specifies where the kernel should connect its console to.
	Console string `yaml:"console"`
	// Timestamp is the the timestamp of installer config creation.
	Timestamp string `yaml:"timestamp"`
	// Nics are the network interfaces of this machine including their neighbors.
	Nics []*apiv2.MachineNic `yaml:"nics"`
	// VPN is the config for connecting machine to VPN
	VPN *apiv2.MachineVPN `yaml:"vpn"`
	// Role is either firewall or machine
	Role apiv2.MachineAllocationType `yaml:"role"`
	// RaidEnabled is set to true if any raid devices are specified
	RaidEnabled bool `yaml:"raidenabled"`
	// RootUUID is the fs uuid if the root fs
	RootUUID string `yaml:"root_uuid"`
	// FirewallRules if not empty firewall rules to enforce
	FirewallRules *apiv2.FirewallRules `yaml:"firewall_rules"`
	// DNSServers for the machine
	DNSServers []*apiv2.DNSServer `yaml:"dns_servers"`
	// NTPServers for the machine
	NTPServers []*apiv2.NTPServer `yaml:"ntp_servers"`
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
