package ubuntu_test

import (
	"log/slog"
	goos "os"
	"testing"

	"github.com/google/go-cmp/cmp"
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/metal-stack/os-installer/pkg/exec"
	oscommon "github.com/metal-stack/os-installer/pkg/installer/os/common"
	"github.com/metal-stack/os-installer/pkg/installer/os/ubuntu"
	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultOS_WriteHostname(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		fsMocks    func(fs *afero.Afero)
		want       string
		wantErr    error
	}{
		{
			name: "write hostname",
			allocation: &apiv2.MachineAllocation{
				Hostname: "test-hostname",
			},
			want:    "test-hostname",
			wantErr: nil,
		},
		{
			name: "overwrite when already exists",
			allocation: &apiv2.MachineAllocation{
				Hostname: "test-hostname",
			},
			want: "test-hostname",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, fs.WriteFile(oscommon.HostnameFilePath, []byte("bar"), goos.ModePerm))
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
				Log:        log,
				Fs:         fs,
				Allocation: tt.allocation,
				Exec:       exec.New(log).WithCommandFn(test.FakeCmd(t)),
			})

			gotErr := d.WriteHostname(t.Context())
			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			content, err := fs.ReadFile(oscommon.HostnameFilePath)
			require.NoError(t, err)

			assert.Equal(t, tt.want, string(content))
		})
	}
}
