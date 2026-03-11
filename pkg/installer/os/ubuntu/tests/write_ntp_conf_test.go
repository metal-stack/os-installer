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
				NtpServer: []*apiv2.NTPServer{
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
		// FIXME!
		// 		{
		// 			name: "configure custom ntp for firewall",
		// 			fsMocks: func(fs *afero.Afero) {
		// 				require.NoError(t, fs.WriteFile(oscommon.ChronyConfigPath, []byte(""), 0644))
		// 			},
		// 			allocation: &apiv2.MachineAllocation{
		// 				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
		// 				NtpServer: []*apiv2.NTPServer{
		// 					{Address: "custom.1.ntp.org"},
		// 					{Address: "custom.2.ntp.org"},
		// 				},
		// 				Project: "project-a",
		// 				Networks: []*apiv2.MachineNetwork{
		// 					{
		// 						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
		// 						Project:     new("project-a"),
		// 						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
		// 						Prefixes:    []string{"10.0.16.0/22"},
		// 						Ips:         []string{"10.0.16.2"},
		// 						Vrf:         3981,
		// 						Asn:         4200003073,
		// 					},
		// 				},
		// 			},
		// 			want: `# Welcome to the chrony configuration file. See chrony.conf(5) for more
		// # information about usable directives.

		// # In case no custom NTP server is provided
		// # Cloudflare offers a free public time service that allows us to use their
		// # anycast network of 180+ locations to synchronize time from their closest server.
		// # See https://blog.cloudflare.com/secure-time/
		// pool custom.1.ntp.org iburst
		// pool custom.2.ntp.org iburst

		// # This directive specify the location of the file containing ID/key pairs for
		// # NTP authentication.
		// keyfile /etc/chrony/chrony.keys

		// # This directive specify the file into which chronyd will store the rate
		// # information.
		// driftfile /var/lib/chrony/chrony.drift

		// # Uncomment the following line to turn logging on.
		// #log tracking measurements statistics

		// # Log files location.
		// logdir /var/log/chrony

		// # Stop bad estimates upsetting machine clock.
		// maxupdateskew 100.0

		// # This directive enables kernel synchronisation (every 11 minutes) of the
		// # real-time clock. Note that it can’t be used along with the 'rtcfile' directive.
		// rtcsync

		// # Step the system clock instead of slewing it if the adjustment is larger than
		// # one second, but only in the first three clock updates.
		// makestep 1 3`,
		// 			wantErr: nil,
		// 		},
		// 		{
		// 			name: "use default ntp for firewall",
		// 			fsMocks: func(fs *afero.Afero) {
		// 				require.NoError(t, fs.WriteFile(oscommon.ChronyConfigPath, []byte(""), 0644))
		// 			},
		// 			allocation: &apiv2.MachineAllocation{
		// 				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
		// 			},
		// 			want:    "",
		// 			wantErr: nil,
		// 		},

		// 		{
		// 			name: "configure ntp for almalinux machine",
		// 			fsMocks: func(fs afero.Fs) {
		// 				require.NoError(t, afero.WriteFile(fs, "/etc/chrony.conf", []byte(""), 0644))
		// 			},
		// 			oss:        osAlmalinux,
		// 			ntpPath:    "/etc/chrony.conf",
		// 			role:       "machine",
		// 			ntpServers: []*models.V1NTPServer{{Address: new("custom.1.ntp.org")}, {Address: new("custom.2.ntp.org")}},
		// 			want: `# Welcome to the chrony configuration file. See chrony.conf(5) for more
		// # information about usable directives.

		// # In case no custom NTP server is provided
		// # Cloudflare offers a free public time service that allows us to use their
		// # anycast network of 180+ locations to synchronize time from their closest server.
		// # See https://blog.cloudflare.com/secure-time/
		// pool custom.1.ntp.org iburst
		// pool custom.2.ntp.org iburst

		// # This directive specify the location of the file containing ID/key pairs for
		// # NTP authentication.
		// keyfile /etc/chrony/chrony.keys

		// # This directive specify the file into which chronyd will store the rate
		// # information.
		// driftfile /var/lib/chrony/chrony.drift

		// # Uncomment the following line to turn logging on.
		// #log tracking measurements statistics

		// # Log files location.
		// logdir /var/log/chrony

		// # Stop bad estimates upsetting machine clock.
		// maxupdateskew 100.0

		// # This directive enables kernel synchronisation (every 11 minutes) of the
		// # real-time clock. Note that it can’t be used along with the 'rtcfile' directive.
		// rtcsync

		// # Step the system clock instead of slewing it if the adjustment is larger than
		// # one second, but only in the first three clock updates.
		// makestep 1 3`,
		// 			wantErr: nil,
		// 		},
		// 		{
		// 			name: "use default ntp for almalinux machine",
		// 			fsMocks: func(fs afero.Fs) {
		// 				require.NoError(t, afero.WriteFile(fs, "/etc/chrony.conf", []byte(""), 0644))
		// 			},
		// 			oss:     osAlmalinux,
		// 			ntpPath: "/etc/chrony.conf",
		// 			role:    "machine",
		// 			want:    "",
		// 			wantErr: nil,
		// 		},
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

			if tt.allocation.AllocationType == apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL {
				content, err := fs.ReadFile(oscommon.ChronyConfigPath)
				require.NoError(t, err)

				assert.Equal(t, tt.want, string(content))

				return
			}

			content, err := fs.ReadFile(oscommon.TimesyncdConfigPath)
			require.NoError(t, err)

			assert.Equal(t, tt.want, string(content))
		})
	}
}
