package installer

import (
	"context"
	"fmt"
	"os"

	"buf.build/go/protoyaml"
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	v1 "github.com/metal-stack/os-installer/api/v1"
	"go.yaml.in/yaml/v3"
)

// ReadConfigurations returns the configuration that were provided from the metal-hammer, which
// were persisted as configuration files on the disk.
// The installer must have run before calling this function, otherwise the files are not there!
func ReadConfigurations() (*v1.MachineDetails, *apiv2.MachineAllocation, error) {
	data, err := os.ReadFile(v1.MachineDetailsPath)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read machine details: %w", err)
	}

	var details v1.MachineDetails
	if err = yaml.Unmarshal(data, &details); err != nil {
		return nil, nil, fmt.Errorf("unable to parse machine details: %w", err)
	}

	data, err = os.ReadFile(v1.MachineAllocationPath)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read machine allocation: %w", err)
	}

	var allocation apiv2.MachineAllocation
	if err = protoyaml.Unmarshal(data, &allocation); err != nil {
		return nil, nil, fmt.Errorf("unable to parse machine allocation: %w", err)
	}

	return &details, &allocation, nil
}

// persistConfigurations writes the configuration data provided from the metal-hammer to the os.
// these can be used again for other applications like the firewall-controller at a later point in time.
func (i *installer) persistConfigurations(context.Context) error {
	detailsBytes, err := yaml.Marshal(i.details)
	if err != nil {
		return fmt.Errorf("unable to marshal machine details: %w", err)
	}

	err = i.fs.WriteFile(v1.MachineDetailsPath, detailsBytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to persist machine details: %w", err)
	}

	allocationBytes, err := protoyaml.Marshal(i.allocation)
	if err != nil {
		return fmt.Errorf("unable to marshal machine allocation: %w", err)
	}

	err = i.fs.WriteFile(v1.MachineAllocationPath, allocationBytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to persist machine allocation: %w", err)
	}

	return nil
}
