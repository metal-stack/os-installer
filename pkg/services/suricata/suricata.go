package suricata

import (
	"context"
	"log/slog"

	systemd_renderer "github.com/metal-stack/os-installer/pkg/systemd-service-renderer"
	renderer "github.com/metal-stack/os-installer/pkg/template-renderer"
	"github.com/spf13/afero"

	_ "embed"
)

const (
	suricataServiceName = "suricata.service"

	suricataUpdateServiceName     = "suricata_update.service"
	suricataUpdateServiceUnitPath = "/etc/systemd/system/" + suricataUpdateServiceName

	suricataDefaultsPath = "/etc/default/suricata"
	suricataConfigPath   = "/etc/suricata/suricata.yaml"
)

var (
	//go:embed suricata.yaml.tpl
	suricataConfigTemplateString string
	//go:embed suricata_defaults.tpl
	suricataDefaultsTemplateString string
	//go:embed suricata_update.service.tpl
	suricataUpdateServiceTemplateString string
)

type Config struct {
	Log    *slog.Logger
	Enable bool
	Reload bool
	fs     afero.Fs
}

type TemplateData struct {
	Interface       string
	DefaultRouteVrf string
}

func WriteSystemdUnit(ctx context.Context, cfg *Config, c *TemplateData) (changed bool, err error) {
	r, err := systemd_renderer.New(&systemd_renderer.Config{
		ServiceName:    suricataUpdateServiceName,
		Log:            cfg.Log,
		TemplateString: suricataUpdateServiceTemplateString,
		Data:           c,
		Fs:             cfg.fs,
	})
	if err != nil {
		return false, err
	}

	chg, err := r.Render(ctx, suricataUpdateServiceUnitPath, cfg.Reload)
	if err != nil {
		return chg, err
	}

	// return changed if one has changed
	changed = changed || chg

	for _, spec := range []struct {
		path           string
		templateString string
	}{
		{
			path:           suricataDefaultsPath,
			templateString: suricataDefaultsTemplateString,
		},
		{
			path:           suricataConfigPath,
			templateString: suricataConfigTemplateString,
		},
	} {
		r, err := renderer.New(&renderer.Config{
			Log:            cfg.Log,
			TemplateString: spec.templateString,
			Data:           c,
			Fs:             cfg.fs,
		})
		if err != nil {
			return false, err
		}

		chg, err := r.Render(ctx, spec.path)
		if err != nil {
			return chg, err
		}

		changed = changed || chg
	}

	if cfg.Reload && changed {
		if err := systemd_renderer.Reload(ctx, cfg.Log.With("service-name", "suricata"), suricataServiceName); err != nil {
			return changed, err
		}
	}

	return changed, nil
}
