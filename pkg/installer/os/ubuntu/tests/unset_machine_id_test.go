package ubuntu_test

import (
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/os-installer/pkg/exec"
	oscommon "github.com/metal-stack/os-installer/pkg/installer/os/common"
	"github.com/metal-stack/os-installer/pkg/installer/os/ubuntu"
	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_os_UnsetMachineID(t *testing.T) {
	tests := []struct {
		name    string
		fsMocks func(fs *afero.Afero)
		wantErr error
	}{
		{
			name: "unset",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, fs.WriteFile("/etc/machine-id", []byte("uuid"), 0700))
				require.NoError(t, fs.WriteFile("/var/lib/dbus/machine-id", []byte("uuid"), 0700))
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
				Log:  log,
				Fs:   fs,
				Exec: exec.New(log).WithCommandFn(test.FakeCmd(t)),
			})

			content, err := fs.ReadFile(oscommon.EtcMachineID)
			require.NoError(t, err)
			require.Equal(t, "uuid", string(content))

			content, err = fs.ReadFile(oscommon.DbusMachineID)
			require.NoError(t, err)
			require.Equal(t, "uuid", string(content))

			gotErr := d.UnsetMachineID(t.Context())
			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			content, err = fs.ReadFile(oscommon.EtcMachineID)
			require.NoError(t, err)
			assert.Empty(t, content)

			content, err = fs.ReadFile(oscommon.DbusMachineID)
			require.NoError(t, err)
			assert.Empty(t, content)
		})
	}
}
