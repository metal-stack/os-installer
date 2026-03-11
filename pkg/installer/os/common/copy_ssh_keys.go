package oscommon

import (
	"context"
	"os/user"
	"path"
	"strconv"
	"strings"
)

func (d *DefaultOS) CopySSHKeys(ctx context.Context) error {
	var (
		sshPath               = path.Join("/home", metalUser, ".ssh")
		sshAuthorizedKeysPath = path.Join(sshPath, "authorized_keys")
	)

	err := d.fs.MkdirAll(sshPath, 0700)
	if err != nil {
		return err
	}

	u, err := user.Lookup(metalUser)
	if err != nil {
		return err
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return err
	}

	err = d.fs.Chown(sshPath, uid, gid)
	if err != nil {
		return err
	}

	var lines []string
	for _, key := range d.allocation.SshPublicKeys {
		lines = append(lines, key)
	}

	err = d.fs.WriteFile(sshAuthorizedKeysPath, []byte(strings.Join(lines, "\n")), 0600)
	if err != nil {
		return err
	}

	return d.fs.Chown(sshAuthorizedKeysPath, uid, gid)
}
