package oscommon

import (
	"context"
	"strings"

	"github.com/spf13/afero"
)

const (
	ResolvConfPath = "/etc/resolv.conf"
)

func (d *CommonTasks) WriteResolvConf(ctx context.Context) error {
	d.log.Info("write configuration", "file", ResolvConfPath)
	// Must be written here because during docker build this file is synthetic
	err := d.fs.Remove(ResolvConfPath)
	if err != nil {
		d.log.Info("config file not present", "file", ResolvConfPath)
	}

	content := []byte(
		`nameserver 8.8.8.8
nameserver 8.8.4.4
`)

	if len(d.allocation.DnsServer) > 0 {
		var s strings.Builder
		for _, dnsServer := range d.allocation.DnsServer {
			s.WriteString("nameserver " + dnsServer.Ip + "\n")
		}

		content = []byte(s.String())
	}

	return afero.WriteFile(d.fs, ResolvConfPath, content, 0644)
}
