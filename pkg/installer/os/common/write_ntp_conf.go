package oscommon

import (
	"context"
	"fmt"
	"strings"
)

const (
	TimesyncdConfigPath = "/etc/systemd/timesyncd.conf"
	ChronyConfigPath    = "/etc/chrony/chrony.conf"
)

func (d *CommonTasks) WriteNTPConf(ctx context.Context) error {
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
