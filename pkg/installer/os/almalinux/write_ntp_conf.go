package almalinux

import (
	"context"
	"fmt"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
)

const (
	chronyConfigPath = "/etc/chrony.conf"
)

func (o *Os) WriteNTPConf(ctx context.Context) error {
	if len(o.allocation.NtpServers) == 0 {
		return nil
	}

	var ntpServers []string

	for _, ntp := range o.allocation.NtpServers {
		ntpServers = append(ntpServers, ntp.Address)
	}

	if o.allocation.AllocationType == apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL {
		return fmt.Errorf("almalinux as firewall is currently not supported")
	}

	return o.WriteNtpConfToPath(chronyConfigPath, ntpServers)
}
