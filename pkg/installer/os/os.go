package operatingsystem

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/metal-stack/os-installer/pkg/exec"
	"github.com/metal-stack/os-installer/pkg/installer/os/almalinux"
	oscommon "github.com/metal-stack/os-installer/pkg/installer/os/common"
	"github.com/metal-stack/os-installer/pkg/installer/os/debian"
	"github.com/metal-stack/os-installer/pkg/installer/os/ubuntu"
	"github.com/spf13/afero"
)

const (
	ubuntuOS    = osName("ubuntu")
	debianOS    = osName("debian")
	almalinuxOS = osName("almalinux")
	// defaultOS contains no specific overwrites and can be used by out-of-tree images
	defaultOS = osName("default")
)

type (
	osName string
)

func New(cfg *oscommon.Config) (oscommon.OperatingSystem, error) {
	if cfg.Log == nil {
		return nil, fmt.Errorf("log must be passed to os-installer")
	}
	if cfg.Allocation == nil {
		return nil, fmt.Errorf("allocation must be passed to os-installer")
	}
	if cfg.MachineDetails == nil {
		return nil, fmt.Errorf("machine details must be passed to os-installer")
	}

	if cfg.Fs == nil {
		cfg.Fs = &afero.Afero{
			Fs: afero.OsFs{},
		}
	}

	if cfg.Exec == nil {
		cfg.Exec = exec.New(cfg.Log)
	}

	if cfg.Name != nil {
		return fromOsName(*cfg.Name, cfg)
	}

	os, err := detectOS(cfg)
	if err != nil {
		cfg.Log.Error("unable to detect operating system, falling back to default implementation", "error", err)
		return fromOsName(string(defaultOS), cfg)
	}

	return os, nil
}

func detectOS(cfg *oscommon.Config) (oscommon.OperatingSystem, error) {
	cfg.Log.Info("automatically detecting operating system for installation")

	content, err := cfg.Fs.ReadFile("/etc/os-release")
	if err != nil {
		return nil, err
	}

	env := map[string]string{}
	for line := range strings.SplitSeq(string(content), "\n") {
		k, v, found := strings.Cut(line, "=")
		if found {
			env[k] = v
		}
	}

	if os, ok := env["ID"]; ok {
		unquoted, err := strconv.Unquote(os)
		if err == nil {
			os = unquoted
		}

		return fromOsName(os, cfg)
	}

	return nil, fmt.Errorf("unable to detect os, no ID field found /etc/os-release")
}

func fromOsName(name string, cfg *oscommon.Config) (oscommon.OperatingSystem, error) {
	switch os := osName(strings.ToLower(name)); os {
	case ubuntuOS:
		cfg.Log.Info("using ubuntu os-installer")
		return ubuntu.New(cfg), nil
	case debianOS:
		cfg.Log.Info("using debian os-installer")
		return debian.New(cfg), nil
	case almalinuxOS:
		cfg.Log.Info("using almalinux os-installer")
		return almalinux.New(cfg), nil
	default:
		cfg.Log.Info("using default os-installer implementation")
		return oscommon.NewDefaultOS(cfg), nil
	}
}
