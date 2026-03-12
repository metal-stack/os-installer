package debian

import (
	"context"

	oscommon "github.com/metal-stack/os-installer/pkg/installer/os/common"
)

type (
	os struct {
		*oscommon.CommonTasks
	}
)

func New(cfg *oscommon.Config) *os {
	return &os{
		CommonTasks: oscommon.New(cfg),
	}
}

func (o *os) BootloaderID() string {
	return "metal-debian"
}

func (o *os) WriteBootInfo(ctx context.Context, cmdLine string) error {
	return o.CommonTasks.WriteBootInfo(ctx, o.InitramdiskFormatString(), o.BootloaderID(), cmdLine)
}

func (o *os) CreateMetalUser(ctx context.Context) error {
	return o.CommonTasks.CreateMetalUser(ctx, o.SudoGroup())
}

func (o *os) GrubInstall(ctx context.Context, cmdLine string) error {
	return o.CommonTasks.GrubInstall(ctx, o.BootloaderID(), cmdLine)
}
