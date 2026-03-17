package oscommon

import "context"

type (
	DefaultOS struct {
		*CommonTasks
		bootloaderID *string
	}
)

func NewDefaultOS(cfg *Config) *DefaultOS {
	return &DefaultOS{
		CommonTasks:  New(cfg),
		bootloaderID: cfg.BootloaderID,
	}
}

func (o *DefaultOS) BootloaderID() string {
	if o.bootloaderID == nil {
		panic("no bootloader id provided for default os")
	}

	return *o.bootloaderID
}

func (o *DefaultOS) WriteBootInfo(ctx context.Context, cmdLine string) error {
	return o.CommonTasks.WriteBootInfo(ctx, o.InitramdiskFormatString(), o.BootloaderID(), cmdLine)
}

func (o *DefaultOS) CreateMetalUser(ctx context.Context) error {
	return o.CommonTasks.CreateMetalUser(ctx, o.SudoGroup())
}

func (o *DefaultOS) GrubInstall(ctx context.Context, cmdLine string) error {
	return o.CommonTasks.GrubInstall(ctx, o.BootloaderID(), cmdLine)
}
