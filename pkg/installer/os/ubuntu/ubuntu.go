package ubuntu

import (
	"context"

	oscommon "github.com/metal-stack/os-installer/pkg/installer/os/common"
)

type (
	Os struct {
		*oscommon.CommonTasks
	}
)

func New(cfg *oscommon.Config) *Os {
	return &Os{
		CommonTasks: oscommon.New(cfg),
	}
}

func (o *Os) BootloaderID() string {
	return "metal-ubuntu"
}

func (o *Os) WriteBootInfo(ctx context.Context, cmdLine string) error {
	return o.CommonTasks.WriteBootInfo(ctx, o.InitramdiskFormatString(), o.BootloaderID(), cmdLine)
}

func (o *Os) CreateMetalUser(ctx context.Context) error {
	return o.CommonTasks.CreateMetalUser(ctx, o.SudoGroup())
}

func (o *Os) GrubInstall(ctx context.Context, cmdLine string) error {
	return o.CommonTasks.GrubInstall(ctx, o.BootloaderID(), cmdLine)
}
