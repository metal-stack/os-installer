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

		switch os := osName(strings.ToLower(os)); os {
		case ubuntuOS:
			return ubuntu.New(cfg), nil
		case debianOS:
			return debian.New(cfg), nil
		case almalinuxOS:
			return almalinux.New(cfg), nil
		default:
			return nil, fmt.Errorf("unsupported operating system: %s", os)
		}
	}

	return nil, fmt.Errorf("unable to detect OS")
}
