package almalinux_test

import (
	"log/slog"
	"os/user"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "github.com/metal-stack/os-installer/api/v1"
	"github.com/metal-stack/os-installer/pkg/exec"
	"github.com/metal-stack/os-installer/pkg/installer/os/almalinux"
	oscommon "github.com/metal-stack/os-installer/pkg/installer/os/common"
	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/spf13/afero"
)

func Test_os_CreateMetalUser(t *testing.T) {
	tests := []struct {
		name         string
		details      *v1.MachineDetails
		execMocks    []test.FakeExecParams
		lookupUserFn oscommon.LookupUserFn
		want         string
		wantErr      error
	}{
		{
			name: "create user already exists",
			details: &v1.MachineDetails{
				Password: "abc",
			},
			lookupUserFn: func(name string) (*user.User, error) {
				return &user.User{
					Uid:      "1000",
					Gid:      "1000",
					Username: oscommon.MetalUser,
					Name:     oscommon.MetalUser,
					HomeDir:  "/home/metal",
				}, nil
			},
			execMocks: []test.FakeExecParams{
				{
					WantCmd:  []string{"userdel", oscommon.MetalUser},
					Output:   "",
					ExitCode: 0,
				},
				{
					WantCmd:  []string{"useradd", "--create-home", "--uid", "1000", "--gid", "wheel", "--shell", "/bin/bash", oscommon.MetalUser},
					Output:   "",
					ExitCode: 0,
				},
				{
					WantCmd:  []string{"passwd", oscommon.MetalUser},
					Output:   "",
					ExitCode: 0,
				},
				{
					WantCmd:  []string{"passwd", "root"},
					Output:   "",
					ExitCode: 0,
				},
			},
		},
		{
			name: "create user does not yet exist",
			details: &v1.MachineDetails{
				Password: "abc",
			},
			lookupUserFn: func(name string) (*user.User, error) {
				return nil, user.UnknownUserError(oscommon.MetalUser)
			},
			execMocks: []test.FakeExecParams{
				{
					WantCmd:  []string{"useradd", "--create-home", "--uid", "1000", "--gid", "wheel", "--shell", "/bin/bash", oscommon.MetalUser},
					Output:   "",
					ExitCode: 0,
				},
				{
					WantCmd:  []string{"passwd", oscommon.MetalUser},
					Output:   "",
					ExitCode: 0,
				},
				{
					WantCmd:  []string{"passwd", "root"},
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

			d := almalinux.New(&oscommon.Config{
				Log:            log,
				Fs:             fs,
				Exec:           exec.New(log).WithCommandFn(test.FakeCmd(t, tt.execMocks...)),
				MachineDetails: tt.details,
				LookupUserFn:   tt.lookupUserFn,
			})

			gotErr := d.CreateMetalUser(t.Context())
			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}
		})
	}
}
