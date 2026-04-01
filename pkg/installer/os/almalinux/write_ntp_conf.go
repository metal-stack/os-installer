package almalinux

import (
	"context"
	"fmt"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/metal-stack/os-installer/pkg/services/chrony"
	renderer "github.com/metal-stack/os-installer/pkg/template-renderer"
)

const (
	ChronyConfigPath = "/etc/chrony.conf"
)

func (o *Os) WriteNTPConf(ctx context.Context) error {
	if o.allocation.AllocationType == apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL {
		return fmt.Errorf("almalinux as firewall is currently not supported")
	}

	if len(o.allocation.NtpServers) == 0 {
		return nil
	}

	var ntpServers []string

	for _, ntp := range o.allocation.NtpServers {
		ntpServers = append(ntpServers, ntp.Address)
	}

	r, err := renderer.New(&renderer.Config{
		Log:            o.log,
		TemplateString: chrony.ChronyConfigTemplateString,
		Data: chrony.TemplateData{
			NTPServers: ntpServers,
		},
		Fs: o.fs,
	})
	if err != nil {
		return err
	}

	_, err = r.Render(ctx, ChronyConfigPath)
	return err
}
