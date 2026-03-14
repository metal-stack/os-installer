package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/metal-stack/os-installer/pkg/installer"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	details, allocation, err := installer.ReadConfigurations()
	if err != nil {
		log.Error("unable to read configuration", "error", err)
	}

	i := installer.New(log, details, allocation)

	if err := i.Install(context.Background()); err != nil {
		log.Error("error while running machine installer", "error", err)
		os.Exit(1)
	}
}
