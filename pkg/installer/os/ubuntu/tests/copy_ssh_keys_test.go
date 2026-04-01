package ubuntu_test

import (
	"log/slog"
	"os/user"
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

func Test_os_CopySSHKeys(t *testing.T) {
	tests := []struct {
		name         string
		allocation   *apiv2.MachineAllocation
		lookupUserFn oscommon.LookupUserFn
		wantErr      error
	}{
		{
			name: "copy ssh keys",
			lookupUserFn: func(name string) (*user.User, error) {
				return &user.User{
					Uid:      "1000",
					Gid:      "1000",
					Username: oscommon.MetalUser,
					Name:     oscommon.MetalUser,
					HomeDir:  "/home/metal",
				}, nil
			},
			allocation: &apiv2.MachineAllocation{
				SshPublicKeys: []string{"a", "b"},
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

			d := ubuntu.New(&oscommon.Config{
				Log:          log,
				Fs:           fs,
				Exec:         exec.New(log).WithCommandFn(test.FakeCmd(t)),
				LookupUserFn: tt.lookupUserFn,
				Allocation:   tt.allocation,
			})

			gotErr := d.CopySSHKeys(t.Context())
			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			content, err := fs.ReadFile("/home/metal/.ssh/authorized_keys")
			require.NoError(t, err)

			assert.Equal(t, "a\nb", string(content))
		})
	}
}
