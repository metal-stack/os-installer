package nftablesexporter

import (
	"context"
	_ "embed"
	"log/slog"

	renderer "github.com/metal-stack/os-installer/pkg/template-renderer"
	"github.com/spf13/afero"
)

const (
	nftablesExporterServiceName     = "nftables-exporter.service"
	nftablesExporterServiceUnitPath = "/etc/systemd/system/" + nftablesExporterServiceName
)

var (
	//go:embed nftables_exporter.service.tpl
	nftablesExporterTemplateString string
)

type DroptailerConfig struct {
	Log    *slog.Logger
	Reload bool
	fs     afero.Fs
}

type DroptailerTemplateData struct {
	Comment   string
}

func WriteSystemdUnit(ctx context.Context, cfg *DroptailerConfig, c *DroptailerTemplateData) (changed bool, err error) {
	r, err := renderer.New(&renderer.Config{
		Log:            cfg.Log,
		ServiceName:    nftablesExporterServiceName,
		TemplateString: nftablesExporterTemplateString,
		Data:           c,
		Fs:             cfg.fs,
	})
	if err != nil {
		return false, err
	}

	return r.Render(ctx, nftablesExporterServiceUnitPath, cfg.Reload)
}
