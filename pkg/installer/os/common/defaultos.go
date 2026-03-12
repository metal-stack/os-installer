package oscommon

import "context"

type (
	defaultOS struct {
		*CommonTasks
		bootloaderID *string
	}
)

func NewDefaultOS(cfg *Config) *defaultOS {
	return &defaultOS{
		CommonTasks:  New(cfg),
		bootloaderID: cfg.BootloaderID,
	}
}

func (o *defaultOS) BootloaderID() string {
	if o.bootloaderID == nil {
		panic("no bootloader id provided for default os")
	}

	return *o.bootloaderID
}

func (o *defaultOS) WriteBootInfo(ctx context.Context, cmdLine string) error {
	return o.CommonTasks.WriteBootInfo(ctx, o.InitramdiskFormatString(), o.BootloaderID(), cmdLine)
}

func (o *defaultOS) CreateMetalUser(ctx context.Context) error {
	return o.CommonTasks.CreateMetalUser(ctx, o.SudoGroup())
}

func (o *defaultOS) GrubInstall(ctx context.Context, cmdLine string) error {
	return o.CommonTasks.GrubInstall(ctx, o.BootloaderID(), cmdLine)
}
