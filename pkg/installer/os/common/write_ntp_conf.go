package oscommon

import (
	"context"
	"fmt"
	"strings"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
)

const (
	TimesyncdConfigPath = "/etc/systemd/timesyncd.conf"
)

func (d *CommonTasks) WriteNTPConf(ctx context.Context) error {
	if d.allocation.AllocationType == apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL {
		d.log.Info("skipping timesyncd config for firewalls as chrony will be configured later on through systemd service renderer")
		return nil
	}

	if len(d.allocation.NtpServers) == 0 {
		return nil
	}

	var ntpServers []string

	for _, ntp := range d.allocation.NtpServers {
		ntpServers = append(ntpServers, ntp.Address)
	}

	return d.WriteNtpConfToPath(TimesyncdConfigPath, ntpServers)
}

func (d *CommonTasks) WriteNtpConfToPath(configPath string, ntpServers []string) error {
	content := fmt.Sprintf("[Time]\nNTP=%s\n", strings.Join(ntpServers, " "))

	err := d.fs.Remove(configPath)
	if err != nil {
		d.log.Info("ntp config file not present", "file", configPath)
	}

	return d.fs.WriteFile(configPath, []byte(content), 0644)
}
