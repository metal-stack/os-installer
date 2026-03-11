package oscommon

import (
	"context"
	"strings"

	"github.com/metal-stack/os-installer/pkg/exec"

	ignitionConfig "github.com/flatcar/ignition/config/v2_4"
)

const (
	UserdataPath         = "/etc/metal/userdata"
	ignitionUserdataPath = "/etc/metal/config.ign"
)

func (d *DefaultOS) ProcessUserdata(ctx context.Context) error {
	if ok := d.fileExists(UserdataPath); !ok {
		d.log.Info("no userdata present, not processing userdata", "path", UserdataPath)
		return nil
	}

	content, err := d.fs.ReadFile(UserdataPath)
	if err != nil {
		return err
	}

	defer func() {
		out, err := d.exec.Execute(ctx, &exec.Params{
			Name: "systemctl",
			Args: []string{"preset-all"},
		})
		if err != nil {
			d.log.Error("error when running systemctl preset-all, continuing anyway", "error", err, "output", string(out))
		}
	}()

	if isCloudInitFile(content) {
		_, err := d.exec.Execute(ctx, &exec.Params{
			Name: "cloud-init",
			Args: []string{"devel", "schema", "--config-file", UserdataPath},
		})
		if err != nil {
			d.log.Error("error when running cloud-init userdata, continuing anyway", "error", err)
		}

		return nil
	}

	err = d.fs.Rename(UserdataPath, ignitionUserdataPath)
	if err != nil {
		return err
	}

	rawConfig, err := d.fs.ReadFile(ignitionUserdataPath)
	if err != nil {
		return err
	}
	_, report, err := ignitionConfig.Parse(rawConfig)
	if err != nil {
		d.log.Error("error when validating ignition userdata, continuing anyway", "error", err)
	}

	d.log.Info("executing ignition")

	_, err = d.exec.Execute(ctx, &exec.Params{
		Name: "ignition",
		Args: []string{"-oem", "file", "-stage", "files", "-log-to-stdout"},
	})
	if err != nil {
		d.log.Error("error when running ignition, continuing anyway", "report", report.Entries, "error", err)
	}

	return nil
}

func isCloudInitFile(content []byte) bool {
	for i, line := range strings.Split(string(content), "\n") {
		if strings.Contains(line, "#cloud-config") {
			return true
		}
		if i > 1 {
			return false
		}
	}
	return false
}
