package oscommon

import "context"

const (
	EtcMachineID  = "/etc/machine-id"
	DbusMachineID = "/var/lib/dbus/machine-id"
)

func (d *DefaultOS) UnsetMachineID(ctx context.Context) error {
	for _, filePath := range []string{EtcMachineID, DbusMachineID} {
		if !d.fileExists(filePath) {
			continue
		}

		f, err := d.fs.Create(filePath)
		if err != nil {
			return err
		}

		_ = f.Close()
	}

	return nil
}
