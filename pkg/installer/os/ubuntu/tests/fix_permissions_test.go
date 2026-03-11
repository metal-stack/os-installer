package ubuntu_test

import (
	iofs "io/fs"
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

func TestDefaultOS_FixPermissions(t *testing.T) {
	tests := []struct {
		name    string
		fsMocks func(fs afero.Fs)
		wantErr error
	}{
		{
			name: "fix permissions",
			fsMocks: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/var/tmp", 0000))
			},
			wantErr: nil,
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

			if tt.fsMocks != nil {
				tt.fsMocks(fs)
			}

			err := d.FixPermissions(t.Context())
			if diff := cmp.Diff(tt.wantErr, err, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n %s", diff)
			}

			info, err := fs.Stat("/var/tmp")
			require.NoError(t, err)
			assert.Equal(t, iofs.FileMode(01777).Perm(), info.Mode().Perm())
		})
	}
}
