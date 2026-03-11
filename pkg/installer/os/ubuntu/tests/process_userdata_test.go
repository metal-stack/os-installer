package ubuntu_test

import (
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "github.com/metal-stack/os-installer/api/v1"
	"github.com/metal-stack/os-installer/pkg/exec"
	oscommon "github.com/metal-stack/os-installer/pkg/installer/os/common"
	"github.com/metal-stack/os-installer/pkg/installer/os/ubuntu"
	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

const (
	sampleCloudInit = `#cloud-config
# Add groups to the system
# The following example adds the ubuntu group with members 'root' and 'sys'
# and the empty group cloud-users.
groups:
	- admingroup: [root,sys]
	- cloud-users`
	sampleIgnition = `{"ignition":{"config":{},"security":{"tls":{}},"timeouts":{},"version":"2.2.0"}}`
)

func TestDefaultOS_ProcessUserdata(t *testing.T) {
	tests := []struct {
		name      string
		details   *v1.MachineDetails
		fsMocks   func(fs *afero.Afero)
		execMocks []test.FakeExecParams
		want      string
		wantErr   error
	}{
		{
			name: "no userdata given",
		},
		{
			name: "cloud-init",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, afero.WriteFile(fs, oscommon.UserdataPath, []byte(sampleCloudInit), 0700))
			},
			execMocks: []test.FakeExecParams{
				{
					WantCmd:  []string{"cloud-init", "devel", "schema", "--config-file", oscommon.UserdataPath},
					Output:   "",
					ExitCode: 0,
				},
				{
					WantCmd:  []string{"systemctl", "preset-all"},
					Output:   "",
					ExitCode: 0,
				},
			},
		},
		{
			name: "ignition",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, afero.WriteFile(fs, oscommon.UserdataPath, []byte(sampleIgnition), 0700))
			},
			execMocks: []test.FakeExecParams{
				{
					WantCmd:  []string{"ignition", "-oem", "file", "-stage", "files", "-log-to-stdout"},
					Output:   "",
					ExitCode: 0,
				},
				{
					WantCmd:  []string{"systemctl", "preset-all"},
					Output:   "",
					ExitCode: 0,
				},
			},
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

			if tt.fsMocks != nil {
				tt.fsMocks(fs)
			}

			d := ubuntu.New(&oscommon.Config{
				Log:            log,
				Fs:             fs,
				Exec:           exec.New(log).WithCommandFn(test.FakeCmd(t, tt.execMocks...)),
				MachineDetails: tt.details,
			})

			gotErr := d.ProcessUserdata(t.Context())
			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}
		})
	}
}
