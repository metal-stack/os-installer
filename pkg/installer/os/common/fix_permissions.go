package oscommon

import (
	"context"
	"io/fs"
)

func (d *CommonTasks) FixPermissions(ctx context.Context) error {
	for p, perm := range map[string]fs.FileMode{
		"/var/tmp": 01777,
	} {
		err := d.fs.Chmod(p, perm)
		if err != nil {
			return err
		}
	}

	return nil
}
