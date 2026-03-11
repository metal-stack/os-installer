package ubuntu_test

import (
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/metal-stack/os-installer/pkg/exec"
	oscommon "github.com/metal-stack/os-installer/pkg/installer/os/common"
	"github.com/metal-stack/os-installer/pkg/installer/os/ubuntu"
	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/metal-stack/v"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultOS_WriteBuildMeta(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		execMocks  []test.FakeExecParams
		want       string
		wantErr    error
	}{
		{
			name: "build meta gets written",
			execMocks: []test.FakeExecParams{
				{
					WantCmd:  []string{"ignition", "-version"},
					Output:   "Ignition v0.36.2",
					ExitCode: 0,
				},
			},
			want: `---
buildVersion: "456"
buildDate: ""
buildSHA: abc
buildRevision: revision
ignitionVersion: Ignition v0.36.2
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				log = slog.Default()
				fs  = &afero.Afero{
					Fs: afero.NewMemMapFs(),
				}
			)

			d := ubuntu.New(&oscommon.Config{
				Log:        log,
				Fs:         fs,
				Allocation: tt.allocation,
				Exec:       exec.New(log).WithCommandFn(test.FakeCmd(t, tt.execMocks...)),
			})

			v.Version = "456"
			v.GitSHA1 = "abc"
			v.Revision = "revision"

			gotErr := d.WriteBuildMeta(t.Context())
			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			content, err := fs.ReadFile(oscommon.BuildMetaPath)
			require.NoError(t, err)

			assert.Equal(t, tt.want, string(content))
		})
	}
}
