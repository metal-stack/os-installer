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
	eveboxServiceName     = "evebox.service"
	eveboxServiceUnitPath = "/etc/systemd/system/" + eveboxServiceName

	suricataServiceName     = "suricata.service"
	suricataServiceUnitPath = "/etc/systemd/system/" + suricataServiceName
	suricataIpsServiceName  = "suricata-ips.service"
	suricataIpsServicePath  = "/etc/systemd/system/" + suricataIpsServiceName

	suricataUpdateServiceName     = "suricata-update.service"
	suricataUpdateServiceUnitPath = "/etc/systemd/system/" + suricataUpdateServiceName

	suricataDefaultsPath = "/etc/default/suricata"
	suricataConfigPath   = "/etc/suricata/suricata.yaml"
)

var (
	//go:embed evebox.service.tpl
	eveboxServiceTemplateString string
	//go:embed suricata.yaml.tpl
	suricataConfigTemplateString string
	//go:embed suricata_defaults.tpl
	suricataDefaultsTemplateString string
	//go:embed suricata.service.tpl
	suricataServiceTemplateString string
	//go:embed suricata-ips.service.tpl
	suricataIpsServiceTemplateString string
	//go:embed suricata-update.service.tpl
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
	TenantVrf       string
}

func WriteSystemdUnit(ctx context.Context, cfg *Config, c *TemplateData) (changed bool, err error) {
	units := []struct {
		serviceName    string
		path           string
		templateString string
		enable         bool
	}{
		{
			serviceName:    eveboxServiceName,
			path:           eveboxServiceUnitPath,
			templateString: eveboxServiceTemplateString,
			enable:         false,
		},
		{
			serviceName:    suricataUpdateServiceName,
			path:           suricataUpdateServiceUnitPath,
			templateString: suricataUpdateServiceTemplateString,
			enable:         cfg.Enable,
		},
		{
			serviceName:    suricataServiceName,
			path:           suricataServiceUnitPath,
			templateString: suricataServiceTemplateString,
			enable:         cfg.Enable,
		},
		{
			serviceName:    suricataIpsServiceName,
			path:           suricataIpsServicePath,
			templateString: suricataIpsServiceTemplateString,
			enable:         false,
		},
	}

	for _, unit := range units {
		r, err := systemd_renderer.New(&systemd_renderer.Config{
			ServiceName:    unit.serviceName,
			Log:            cfg.Log,
			TemplateString: unit.templateString,
			Data:           c,
			Fs:             cfg.fs,
			Enable:         unit.enable,
		})
		if err != nil {
			return false, err
		}

		chg, err := r.Render(ctx, unit.path, false)
		if err != nil {
			return chg, err
		}

		changed = changed || chg
	}

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
