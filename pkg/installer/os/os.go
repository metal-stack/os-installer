package operatingsystem

import (
	"fmt"
	"strconv"
	"strings"

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
)

type (
	osName string
)

func New(cfg *oscommon.Config) (oscommon.OperatingSystem, error) {
	if cfg.Fs == nil {
		cfg.Fs = &afero.Afero{
			Fs: afero.OsFs{},
		}
	}

	if cfg.Name != nil {
		return fromOsName(*cfg.Name, cfg)
	}

	return detectOS(cfg)
}

func detectOS(cfg *oscommon.Config) (oscommon.OperatingSystem, error) {
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
