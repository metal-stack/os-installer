package chrony

import (
	"context"
	"fmt"
	"log/slog"

	systemd_renderer "github.com/metal-stack/os-installer/pkg/systemd-service-renderer"
	renderer "github.com/metal-stack/os-installer/pkg/template-renderer"
	"github.com/spf13/afero"

	_ "embed"
)

const (
	chronyConfigPath = "/etc/chrony/chrony.conf"
)

var (
	//go:embed chrony.conf.tpl
	ChronyConfigTemplateString string
)

type Config struct {
	Log    *slog.Logger
	Reload bool
	Enable bool
	// ChronyConfigPath allows overwriting the default chrony config path
	ChronyConfigPath string
	fs               afero.Fs
}

type TemplateData struct {
	NTPServers []string
}

func WriteSystemdUnit(ctx context.Context, cfg *Config, c *TemplateData, vrfName string) (changed bool, err error) {
	serviceName := fmt.Sprintf("chrony@%s.service", vrfName)

	r, err := renderer.New(&renderer.Config{
		Log:            cfg.Log,
		TemplateString: ChronyConfigTemplateString,
		Data:           c,
		Fs:             cfg.fs,
	})
	if err != nil {
		return false, err
	}

	path := chronyConfigPath
	if cfg.ChronyConfigPath != "" {
		path = cfg.ChronyConfigPath
	}

	changed, err = r.Render(ctx, path)
	if err != nil {
		return changed, err
	}

	if cfg.Enable {
		if err := systemd_renderer.Enable(ctx, cfg.Log, serviceName); err != nil {
			return changed, err
		}
	}

	if cfg.Reload && changed {
		if err := systemd_renderer.Reload(ctx, cfg.Log.With("service-name", "chrony"), serviceName); err != nil {
			return changed, err
		}
	}

	return changed, nil
}
