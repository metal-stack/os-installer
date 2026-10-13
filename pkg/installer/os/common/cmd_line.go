package oscommon

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/metal-stack/os-installer/pkg/exec"
)

func (d *CommonTasks) BuildCMDLine(ctx context.Context) (string, error) {
	parts := []string{
		fmt.Sprintf("console=%s", d.details.Console),
		fmt.Sprintf("root=UUID=%s", d.details.RootUUID),
		"init=/sbin/init",
		"net.ifnames=0",
		"biosdevname=0",
		"nvme_core.io_timeout=300", // 300 sec should be enough for firewalls to be replaced
	}

	mdUUID, found, err := d.findMDUUID(ctx)
	if err != nil {
		return "", err
	}

	if found {
		mdParts := []string{
			"rdloaddriver=raid0",
			"rdloaddriver=raid1",
			fmt.Sprintf("rd.md.uuid=%s", mdUUID),
		}

		parts = append(parts, mdParts...)
	}

	return strings.Join(parts, " "), nil
}

func (d *CommonTasks) findMDUUID(ctx context.Context) (mdUUID string, found bool, err error) {
	d.log.Debug("detect software raid uuid")

	if !d.details.RaidEnabled {
		return "", false, nil
	}

	blkidOut, err := d.exec.Execute(ctx, &exec.Params{
		Name:    "blkid",
		Timeout: 10 * time.Second,
	})
	if err != nil {
		return "", false, fmt.Errorf("unable to run blkid: %w", err)
	}

	if d.details.RootUUID == "" {
		return "", false, fmt.Errorf("no root uuid set in machine details")
	}

	var (
		rootUUID = d.details.RootUUID
		rootDisk string
	)

	for line := range strings.SplitSeq(string(blkidOut), "\n") {
		if strings.Contains(line, rootUUID) {
			rd, _, found := strings.Cut(line, ":")
			if found {
				rootDisk = strings.TrimSpace(rd)
				break
			}
		}
	}
	if rootDisk == "" {
		return "", false, fmt.Errorf("unable to detect rootdisk")
	}

	mdadmOut, err := d.exec.Execute(ctx, &exec.Params{
		Name:    "mdadm",
		Args:    []string{"--detail", "--export", rootDisk},
		Timeout: 10 * time.Second,
	})
	if err != nil {
		return "", false, fmt.Errorf("unable to run mdadm: %w", err)
	}

	for line := range strings.SplitSeq(string(mdadmOut), "\n") {
		_, md, found := strings.Cut(line, "MD_UUID=")
		if found {
			mdUUID = md
			break
		}
	}

	if mdUUID == "" {
		return "", false, fmt.Errorf("unable to detect md root disk")
	}

	return mdUUID, true, nil
}
