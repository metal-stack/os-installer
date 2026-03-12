package almalinux

import (
	"context"
	"time"

	"github.com/metal-stack/os-installer/pkg/exec"
)

func (o *os) CreateMetalUser(ctx context.Context) error {
	err := o.CommonTasks.CreateMetalUser(ctx, o.SudoGroup())
	if err != nil {
		return err
	}

	// otherwise in rescue mode the root account is locked
	_, err = o.exec.Execute(ctx, &exec.Params{
		Name:    "passwd",
		Args:    []string{"root"},
		Timeout: 10 * time.Second,
		Stdin:   o.details.Password + "\n" + o.details.Password + "\n",
	})
	if err != nil {
		return err
	}

	return nil
}
