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

func Tes_os_WriteResolvConf(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		fsMocks    func(fs *afero.Afero)
		want       string
		wantErr    error
	}{
		{
			name:       "resolv.conf gets written",
			allocation: &apiv2.MachineAllocation{},
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, fs.WriteFile(oscommon.ResolvConfPath, []byte(""), 0755))
			},
			want: `nameserver 8.8.8.8
nameserver 8.8.4.4
`,
			wantErr: nil,
		},
		{
			name:       "resolv.conf gets written, file is not present",
			allocation: &apiv2.MachineAllocation{},
			want: `nameserver 8.8.8.8
nameserver 8.8.4.4
`,
			wantErr: nil,
		},
		{
			name: "overwrite resolv.conf with custom DNS",
			allocation: &apiv2.MachineAllocation{
				DnsServer: []*apiv2.DNSServer{
					{Ip: "1.2.3.4"},
					{Ip: "5.6.7.8"},
				},
			},
			want: `nameserver 1.2.3.4
nameserver 5.6.7.8
`,
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

			gotErr := d.WriteResolvConf(t.Context())
			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			content, err := fs.ReadFile(oscommon.ResolvConfPath)
			require.NoError(t, err)

			assert.Equal(t, tt.want, string(content))
		})
	}
}
