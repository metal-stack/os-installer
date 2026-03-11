package almalinux

import (
	"context"
	"log/slog"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	v1 "github.com/metal-stack/os-installer/api/v1"
	"github.com/metal-stack/os-installer/pkg/exec"
	oscommon "github.com/metal-stack/os-installer/pkg/installer/os/common"
	"github.com/metal-stack/os-installer/pkg/network"
	"github.com/spf13/afero"
)

type (
	os struct {
		*oscommon.DefaultOS
		log        *slog.Logger
		details    *v1.MachineDetails
		allocation *apiv2.MachineAllocation
		exec       *exec.CmdExecutor
		network    *network.Network
		fs         *afero.Afero
	}
)

func New(cfg *oscommon.Config) *os {
	return &os{
		DefaultOS:  oscommon.New(cfg),
		log:        cfg.Log,
		details:    cfg.MachineDetails,
		allocation: cfg.Allocation,
		exec:       cfg.Exec,
		network:    network.New(cfg.Allocation),
		fs:         cfg.Fs,
	}
}

func (o *os) SudoGroup() string {
	return "wheel"
}

func (o *os) BootloaderID() string {
	return "almalinux"
}

func (o *os) InitramdiskFormatString() string {
	return "initramfs-%s.img"
}

func (o *os) WriteBootInfo(ctx context.Context, cmdLine string) error {
	return o.DefaultOS.WriteBootInfo(ctx, o.InitramdiskFormatString(), o.BootloaderID(), cmdLine)
}
