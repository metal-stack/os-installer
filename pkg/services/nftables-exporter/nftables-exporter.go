package nftablesexporter

import (
	"context"
	_ "embed"
	"log/slog"

	renderer "github.com/metal-stack/os-installer/pkg/template-renderer"
	"github.com/spf13/afero"
)

const (
	serviceName     = "nftables-exporter.service"
	serviceUnitPath = "/etc/systemd/system/" + serviceName
)

var (
	//go:embed nftables_exporter.service.tpl
	templateString string
)

type Config struct {
	Log    *slog.Logger
	Reload bool
	fs     afero.Fs
}

type TemplateData struct {
	Comment string
}

func WriteSystemdUnit(ctx context.Context, cfg *Config, c *TemplateData) (changed bool, err error) {
	r, err := renderer.New(&renderer.Config{
		Log:            cfg.Log,
		ServiceName:    serviceName,
		TemplateString: templateString,
		Data:           c,
		Fs:             cfg.fs,
	})
	if err != nil {
		return false, err
	}

	return r.Render(ctx, serviceUnitPath, cfg.Reload)
}
