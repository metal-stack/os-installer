package oscommon

import (
	"context"

	"github.com/metal-stack/os-installer/pkg/services"
)

func (d *CommonTasks) SystemdServices(ctx context.Context) error {
	return services.WriteSystemdServices(ctx, d.log, d.network, d.details.ID)
}
