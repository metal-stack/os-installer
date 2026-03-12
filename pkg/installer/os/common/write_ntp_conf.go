package oscommon

import (
	"context"
	"fmt"
	"strings"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/metal-stack/os-installer/pkg/services/chrony"
)

const (
	TimesyncdConfigPath = "/etc/systemd/timesyncd.conf"
	ChronyConfigPath    = "/etc/chrony/chrony.conf"
)

func (d *DefaultOS) WriteNTPConf(ctx context.Context) error {
	if len(d.allocation.NtpServer) == 0 {
		return nil
	}

	var ntpServers []string

	for _, ntp := range d.allocation.NtpServer {
		ntpServers = append(ntpServers, ntp.Address)
	}

	if d.allocation.AllocationType == apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL {
		defaultVRF, err := d.network.GetTenantNetworkVrfName()
		if err != nil {
			return err
		}

		// TODO: check if this is really required as chrony also gets set up in systemd service task?

		_, err = chrony.WriteSystemdUnit(ctx, &chrony.Config{
			Log:    d.log,
			Reload: false,
			Enable: true,
		}, &chrony.TemplateData{
			NTPServers: ntpServers,
		}, defaultVRF)

		if err != nil {
			return err
		}

		return d.WriteNtpConfToPath(ChronyConfigPath, ntpServers)
	}

	return d.WriteNtpConfToPath(TimesyncdConfigPath, ntpServers)
}

func (d *DefaultOS) WriteNtpConfToPath(configPath string, ntpServers []string) error {
	content := fmt.Sprintf("[Time]\nNTP=%s\n", strings.Join(ntpServers, " "))

	err := d.fs.Remove(configPath)
	if err != nil {
		d.log.Info("ntp config file not present", "file", configPath)
	}

	return d.fs.WriteFile(configPath, []byte(content), 0644)
}
