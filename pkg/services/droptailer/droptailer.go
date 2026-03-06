package droptailer

import (
	"context"
	_ "embed"
	"log/slog"

	renderer "github.com/metal-stack/os-installer/pkg/template-renderer"
	"github.com/spf13/afero"
)

const (
	droptailerServiceName     = "droptailer.service"
	droptailerServiceUnitPath = "/etc/systemd/system/" + droptailerServiceName
)

var (
	//go:embed droptailer.service.tpl
	droptailerTemplateString string
)

type DroptailerConfig struct {
	Log    *slog.Logger
	Reload bool
	fs     afero.Fs
}

type DroptailerTemplateData struct {
	Comment   string
	TenantVrf string
}

func WriteSystemdUnit(ctx context.Context, cfg *DroptailerConfig, c *DroptailerTemplateData) (changed bool, err error) {
	r, err := renderer.New(&renderer.Config{
		Log:            cfg.Log,
		ServiceName:    droptailerServiceName,
		TemplateString: droptailerTemplateString,
		Data:           c,
		Fs:             cfg.fs,
	})
	if err != nil {
		return false, err
	}

	return r.Render(ctx, droptailerServiceUnitPath, cfg.Reload)
}
