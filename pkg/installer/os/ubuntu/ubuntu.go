package ubuntu

import (
	"context"

	oscommon "github.com/metal-stack/os-installer/pkg/installer/os/common"
)

type (
	os struct {
		*oscommon.DefaultOS
	}
)

func New(cfg *oscommon.Config) *os {
	return &os{
		DefaultOS: oscommon.New(cfg),
	}
}

func (o *os) BootloaderID() string {
	return "metal-ubuntu"
}

func (o *os) WriteBootInfo(ctx context.Context, cmdLine string) error {
	return o.DefaultOS.WriteBootInfo(ctx, o.InitramdiskFormatString(), o.BootloaderID(), cmdLine)
}

func (o *os) CreateMetalUser(ctx context.Context) error {
	return o.DefaultOS.CreateMetalUser(ctx, o.SudoGroup())
}

func (o *os) GrubInstall(ctx context.Context, cmdLine string) error {
	return o.DefaultOS.GrubInstall(ctx, o.BootloaderID(), cmdLine)
}
