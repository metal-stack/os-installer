package oscommon

import (
	"context"
	"strings"

	v1 "github.com/metal-stack/os-installer/api/v1"
	"github.com/metal-stack/os-installer/pkg/exec"
	"github.com/metal-stack/v"
	"go.yaml.in/yaml/v3"
)

const (
	BuildMetaPath = "/etc/metal/build-meta.yaml"
)

func (d *CommonTasks) WriteBuildMeta(ctx context.Context) error {
	d.log.Info("writing build meta file", "path", BuildMetaPath)

	meta := &v1.BuildMeta{
		Version:  v.Version,
		Date:     v.BuildDate,
		SHA:      v.GitSHA1,
		Revision: v.Revision,
	}

	out, err := d.exec.Execute(ctx, &exec.Params{
		Name: "ignition",
		Args: []string{"-version"},
	})
	if err != nil {
		d.log.Error("error detecting ignition version for build meta, continuing anyway", "error", err)
	} else {
		meta.IgnitionVersion = strings.TrimSpace(out)
	}

	content, err := yaml.Marshal(meta)
	if err != nil {
		return err
	}

	content = append([]byte("---\n"), content...)

	return d.fs.WriteFile(BuildMetaPath, content, 0644)
}
