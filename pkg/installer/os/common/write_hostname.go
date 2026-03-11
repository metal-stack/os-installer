package oscommon

import (
	"context"
)

const (
	HostnameFilePath = "/etc/hostname"
)

func (d *DefaultOS) WriteHostname(ctx context.Context) error {
	return d.fs.WriteFile(HostnameFilePath, []byte(d.allocation.Hostname), 0644)
}
