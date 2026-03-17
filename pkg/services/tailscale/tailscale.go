package tailscale

import (
	"context"
	"log/slog"

	systemd_renderer "github.com/metal-stack/os-installer/pkg/systemd-service-renderer"
	"github.com/spf13/afero"

	_ "embed"
)

const (
	tailscaleServiceName     = "tailscale.service"
	tailscaleServiceUnitPath = "/etc/systemd/system/" + tailscaleServiceName

	tailscaledServiceName     = "tailscaled.service"
	tailscaledServiceUnitPath = "/etc/systemd/system/" + tailscaledServiceName

	tailscaledDefaultPort = "41641"
)

var (
	//go:embed tailscale.service.tpl
	tailscaleTemplateString string
	//go:embed tailscaled.service.tpl
	tailscaledTemplateString string
)

type Config struct {
	Log    *slog.Logger
	Enable bool
	Reload bool
	fs     afero.Fs
}

type TemplateData struct {
	Comment         string
	DefaultRouteVrf string
	TailscaledPort  string
	MachineID       string
	AuthKey         string
	Address         string
}

func WriteSystemdUnit(ctx context.Context, cfg *Config, c *TemplateData) (changed bool, err error) {
	if c.TailscaledPort == "" {
		c.TailscaledPort = tailscaledDefaultPort
	}

	for _, spec := range []struct {
		servicePath    string
		serviceName    string
		templateString string
	}{
		{
			servicePath:    tailscaleServiceUnitPath,
			serviceName:    tailscaleServiceName,
			templateString: tailscaleTemplateString,
		},
		{
			servicePath:    tailscaledServiceUnitPath,
			serviceName:    tailscaledServiceName,
			templateString: tailscaledTemplateString,
		},
	} {
		r, err := systemd_renderer.New(&systemd_renderer.Config{
			ServiceName:    spec.serviceName,
			Log:            cfg.Log,
			TemplateString: spec.templateString,
			Data:           c,
			Fs:             cfg.fs,
		})
		if err != nil {
			return false, err
		}

		chg, err := r.Render(ctx, spec.servicePath, cfg.Reload)
		if err != nil {
			return chg, err
		}

		// return changed if one has changed
		changed = changed || chg
	}

	return changed, nil
}
