package installer

import (
	"fmt"
	"os"

	v1 "github.com/metal-stack/os-installer/api/v1"
	"go.yaml.in/yaml/v3"
)

func (i *installer) PersistLegacyInstallYaml(installConfig *v1.InstallerConfig) error {
	installBytes, err := yaml.Marshal(installConfig)
	if err != nil {
		return fmt.Errorf("unable to marshal legacy installer config: %w", err)
	}
	err = i.fs.WriteFile(v1.LegacyInstallPath, installBytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to persist legacy installer config: %w", err)
	}
	return nil
}

func ReadLegacyInstallYaml() (*v1.InstallerConfig, error) {
	data, err := os.ReadFile(v1.LegacyInstallPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read legacy installer config: %w", err)
	}

	var installConfig v1.InstallerConfig
	if err = yaml.Unmarshal(data, &installConfig); err != nil {
		return nil, fmt.Errorf("unable to parse legacy installer config: %w", err)
	}

	return &installConfig, nil
}
