package firewallcontroller

import (
	"context"
	"log/slog"

	renderer "github.com/metal-stack/os-installer/pkg/template-renderer"
	"github.com/spf13/afero"

	_ "embed"
)

const (
	firewallControllerServiceName     = "firewall-controller.service"
	firewallControllerServiceUnitPath = "/etc/systemd/system/" + firewallControllerServiceName
)

var (
	//go:embed firewall_controller.service.tpl
	firewallControllerTemplateString string
)

type FirewallControllerConfig struct {
	Log    *slog.Logger
	Reload bool
	fs     afero.Fs
}

type FirewallControllerTemplateData struct {
	Comment         string
	DefaultRouteVrf string
}

func WriteSystemdUnit(ctx context.Context, cfg *FirewallControllerConfig, c *FirewallControllerTemplateData) (changed bool, err error) {
	r, err := renderer.New(&renderer.Config{
		Log:            cfg.Log,
		ServiceName:    firewallControllerServiceName,
		TemplateString: firewallControllerTemplateString,
		Data:           c,
		Fs:             cfg.fs,
	})
	if err != nil {
		return false, err
	}

	return r.Render(ctx, firewallControllerServiceUnitPath, cfg.Reload)
}
