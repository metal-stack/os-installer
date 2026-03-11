package almalinux_test

import (
	"context"
	"fmt"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
)

const (
	chronyConfigPath = "/etc/chrony.conf"
)

func (o *os) WriteNTPConf(ctx context.Context) error {
	if len(o.allocation.NtpServer) == 0 {
		return nil
	}

	var ntpServers []string

	for _, ntp := range o.allocation.NtpServer {
		ntpServers = append(ntpServers, ntp.Address)
	}

	if o.allocation.AllocationType == apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL {
		return fmt.Errorf("almalinux as firewall is currently not supported")
	}

	return o.DefaultOS.WriteNtpConfToPath(chronyConfigPath, ntpServers)
}
