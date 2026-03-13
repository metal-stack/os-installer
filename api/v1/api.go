package v1

import (
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
)

const (
	MachineDetailsPath    = "/etc/metal/machine-details.yaml"
	MachineAllocationPath = "/etc/metal/machine-allocation.yaml"
	InstallerConfigPath   = "/etc/metal/os-installer.yaml"
	LLDPDConfigPath       = "/etc/metal/install.yaml"
	BuildMetaPath         = "/etc/metal/build-meta.yaml"
	BootInfoPath          = "/etc/metal/boot-info.yaml"
)

type (
	// Bootinfo is written by the installer in the target os to tell us
	// which kernel, initrd and cmdline must be used for kexec
	Bootinfo struct {
		Initrd       string `yaml:"initrd"`
		Cmdline      string `yaml:"cmdline"`
		Kernel       string `yaml:"kernel"`
		BootloaderID string `yaml:"bootloader_id"`
	}

	// Config can be placed inside the target OS to customize the os-installer.
	Config struct {
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

	// MachineDetails which are not part of the MachineAllocation but required to complete the installation.
	// Is written by by the metal-hammer
	MachineDetails struct {
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

	// LLDPDConfig contains the configuration which is required for the lldpd to start.
	// must be stored in yaml format at /etc/metal/install.yaml
	// Is written by by the metal-hammer
	LLDPDConfig struct {
		// MachineUUID is the unique UUID for this machine, usually the board serial.
		MachineUUID string `yaml:"machineuuid"`
		// Timestamp is the the timestamp of installer config creation.
		Timestamp string `yaml:"timestamp"`
	}

	// BuildMeta is written after the installation finished to store details about the installation version.
	BuildMeta struct {
		Version  string `json:"buildVersion" yaml:"buildVersion"`
		Date     string `json:"buildDate" yaml:"buildDate"`
		SHA      string `json:"buildSHA" yaml:"buildSHA"`
		Revision string `json:"buildRevision" yaml:"buildRevision"`

		IgnitionVersion string `json:"ignitionVersion" yaml:"ignitionVersion"`
	}
)
