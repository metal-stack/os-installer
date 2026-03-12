package oscommon

import (
	"context"
	"fmt"
	"os/user"
	"time"

	"github.com/metal-stack/os-installer/pkg/exec"
)

const (
	metalUser = "metal"
)

func (d *CommonTasks) CreateMetalUser(ctx context.Context, sudoGroup string) error {
	u, err := user.Lookup(metalUser)
	if err != nil {
		if err.Error() != user.UnknownUserError(metalUser).Error() {
			return err
		}
	}

	if u != nil {
		d.log.Info("user already exists, recreating")

		_, err = d.exec.Execute(ctx, &exec.Params{
			Name:    "userdel",
			Args:    []string{metalUser},
			Timeout: 10 * time.Second,
		})
		if err != nil {
			return err
		}
	}

	_, err = d.exec.Execute(ctx, &exec.Params{
		Name:    "useradd",
		Args:    []string{"--create-home", "--uid", "1000", "--gid", sudoGroup, "--shell", "/bin/bash", metalUser},
		Timeout: 10 * time.Second,
	})
	if err != nil {
		return err
	}

	_, err = d.exec.Execute(ctx, &exec.Params{
		Name:    "passwd",
		Args:    []string{metalUser},
		Timeout: 10 * time.Second,
		Stdin:   d.details.Password + "\n" + d.details.Password + "\n",
	})
	if err != nil {
		return fmt.Errorf("unable to set password for metal user: %w", err)
	}

	return nil
}
