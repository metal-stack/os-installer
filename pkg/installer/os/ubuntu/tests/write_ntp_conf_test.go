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
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_os_WriteNTPConf(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		fsMocks    func(fs *afero.Afero)
		want       string
		wantErr    error
	}{
		{
			name: "configure custom ntp",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, fs.WriteFile(oscommon.TimesyncdConfigPath, []byte(""), 0644))
			},
			allocation: &apiv2.MachineAllocation{
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				NtpServers: []*apiv2.NTPServer{
					{Address: "custom.1.ntp.org"},
					{Address: "custom.2.ntp.org"},
				},
			},
			want: `[Time]
NTP=custom.1.ntp.org custom.2.ntp.org
`,
			wantErr: nil,
		},
		{
			name: "use default ntp",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, fs.WriteFile(oscommon.TimesyncdConfigPath, []byte(""), 0644))
			},
			allocation: &apiv2.MachineAllocation{
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
			},
			want:    "",
			wantErr: nil,
		},
		{
			name: "skip firewalls",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, fs.WriteFile(oscommon.TimesyncdConfigPath, []byte(""), 0644))
			},
			allocation: &apiv2.MachineAllocation{
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				NtpServers: []*apiv2.NTPServer{
					{Address: "custom.1.ntp.org"},
					{Address: "custom.2.ntp.org"},
				},
			},
			want:    "",
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

			gotErr := d.WriteNTPConf(t.Context())
			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			content, err := fs.ReadFile(oscommon.TimesyncdConfigPath)
			require.NoError(t, err)

			assert.Equal(t, tt.want, string(content))
		})
	}
}
