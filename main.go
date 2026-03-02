package main

import (
	"log/slog"
	"os"
	"os/exec"
	"time"

	v1 "github.com/metal-stack/os-installer/api/v1"
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

	config, err := parseInstallYAML(fs)
	if err != nil {
		log.Error("installation failed", "error", err)
		os.Exit(1)
	}

	i := installer{
		log:    log.WithGroup("install-go"),
		fs:     fs,
		oss:    oss,
		config: config,
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

func parseInstallYAML(fs afero.Fs) (*v1.InstallerConfig, error) {
	var config v1.InstallerConfig
	content, err := afero.ReadFile(fs, installYAML)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
