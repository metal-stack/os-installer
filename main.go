package main

import (
	"context"
	"log/slog"
	"os"

	"buf.build/go/protoyaml"
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	v1 "github.com/metal-stack/os-installer/api/v1"
	"github.com/metal-stack/os-installer/pkg/installer"
	"github.com/stretchr/testify/assert/yaml"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	data, err := os.ReadFile(v1.MachineDetailsPath)
	if err != nil {
		log.Error("unable to read machine details", "error", err)
		os.Exit(1)
	}

	var details v1.MachineDetails
	if err = yaml.Unmarshal(data, &details); err != nil {
		log.Error("unable to parse machine details", "error", err)
		os.Exit(1)
	}

	data, err = os.ReadFile(v1.MachineAllocationPath)
	if err != nil {
		log.Error("unable to read machine allocation", "error", err)
		os.Exit(1)
	}

	var allocation apiv2.MachineAllocation
	if err = protoyaml.Unmarshal(data, &allocation); err != nil {
		log.Error("unable to parse machine allocation", "error", err)
		os.Exit(1)
	}

	if err := installer.Install(context.Background(), log, &details, &allocation); err != nil {
		log.Error("error while running machine installer", "error", err)
		os.Exit(1)
	}
}
