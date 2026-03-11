package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/metal-stack/os-installer/pkg/install"
	"github.com/metal-stack/v"
	"github.com/spf13/afero"
)

func main() {
	start := time.Now()
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})
	log := slog.New(jsonHandler)

	log.Info("running install", "version", v.V.String())

	fs := afero.OsFs{}

	config, err := install.ParseInstallYAML(fs)
	if err != nil {
		log.Error("installation failed", "error", err)
		os.Exit(1)
	}

	err = install.Install(log.WithGroup("os-installer"), config)
	if err != nil {
		log.Error("installation failed", "error", err, "duration", time.Since(start).String())
		os.Exit(1)
	}

	log.Info("installation succeeded", "duration", time.Since(start).String())
}
