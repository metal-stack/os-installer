package oscommon

import (
	"context"
	"fmt"

	v1 "github.com/metal-stack/os-installer/api/v1"
	"go.yaml.in/yaml/v3"
)

const (
	BootInfoPath = "/etc/metal/boot-info.yaml"
)

func (d *CommonTasks) WriteBootInfo(ctx context.Context, initramdiskFormatString, bootloaderID, cmdLine string) error {
	kern, initrd, err := d.KernelAndInitrdPath(initramdiskFormatString)
	if err != nil {
		return err
	}

	content, err := yaml.Marshal(v1.Bootinfo{
		Initrd:       initrd,
		Cmdline:      cmdLine,
		Kernel:       kern,
		BootloaderID: bootloaderID,
	})
	if err != nil {
		return fmt.Errorf("unable to write boot-info.yaml: %w", err)
	}

	return d.fs.WriteFile(BootInfoPath, content, 0700)
}
