package main

import (
	"log/slog"
	"os"
	"os/exec"
	"time"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/metal-stack/v"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

func main() {
	start := time.Now()
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})
	log := slog.New(jsonHandler)

	log.Info("running install", "version", v.V.String())

	fs := afero.OsFs{}

	oss, err := detectOS(fs)
	if err != nil {
		log.Error("installation failed", "error", err)
		os.Exit(1)
	}

	config, err := parseAllocationYAML(fs)
	if err != nil {
		log.Error("installation failed", "error", err)
		os.Exit(1)
	}

	i := installer{
		log:   log.WithGroup("os-installer"),
		fs:    fs,
		oss:   oss,
		alloc: config,
		exec: &cmdexec{
			log: log.WithGroup("cmdexec"),
			c:   exec.CommandContext,
		},
	}

	err = i.do()
	if err != nil {
		i.log.Error("installation failed", "error", err, "duration", time.Since(start).String())
		os.Exit(1)
	}

	i.log.Info("installation succeeded", "duration", time.Since(start).String())
}

func parseAllocationYAML(fs afero.Fs) (*apiv2.MachineAllocation, error) {
	var alloc apiv2.MachineAllocation
	content, err := afero.ReadFile(fs, allocationYAML)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(content, &alloc)
	if err != nil {
		return nil, err
	}
	return &alloc, nil
}
