package interfaces

import (
	"embed"
	"log/slog"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/metal-stack/os-installer/pkg/network"
	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	_ "embed"
)

var (
	//go:embed test
	expectedInterfaceFiles embed.FS
)

func Test_configureLoopbackInterface(t *testing.T) {
	tests := []struct {
		name         string
		allocation   *apiv2.MachineAllocation
		wantFilePath string
		wantErr      error
	}{
		{
			name: "render machine",
			allocation: &apiv2.MachineAllocation{
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Ips:         []string{"10.0.17.2"},
					},
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:         []string{"185.1.2.3"},
					},
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:         []string{"100.127.129.1"},
					},
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:         []string{"2001::4"},
					},
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Ips:         []string{"10.1.0.1"},
					},
				},
			},
			wantFilePath: "machine/00-lo.network",
			wantErr:      nil,
		},
		{
			name: "render firewall",
			allocation: &apiv2.MachineAllocation{
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Networks: []*apiv2.MachineNetwork{
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Ips:         []string{"10.0.17.2"},
					},
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:         []string{"185.1.2.3"},
					},
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:         []string{"100.127.129.1"},
					},
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:         []string{"2001::4"},
					},
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Ips:         []string{"10.1.0.1"},
					},
				},
			},
			wantFilePath: "firewall/00-lo.network",
			wantErr:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}

			gotErr := configureLoopbackInterface(t.Context(), &Config{
				Log:     slog.Default(),
				fs:      fs,
				Network: network.New(tt.allocation),
			})

			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			content, err := fs.ReadFile(path.Join(systemdNetworkPath, "00-lo.network"))
			require.NoError(t, err)

			if diff := cmp.Diff(mustReadExpected(tt.wantFilePath), string(content)); diff != "" {
				t.Errorf("diff (+got -want):\n%s", diff)
			}
		})
	}
}

func Test_configureLanInterface(t *testing.T) {
	tests := []struct {
		name          string
		allocation    *apiv2.MachineAllocation
		nics          []*apiv2.MachineNic
		wantFilePaths []string
		wantErr       error
	}{
		{
			name: "render machine",
			allocation: &apiv2.MachineAllocation{
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Ips:         []string{"10.0.17.2"},
					},
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:         []string{"185.1.2.3"},
					},
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:         []string{"100.127.129.1"},
					},
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:         []string{"2001::4"},
					},
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Ips:         []string{"10.1.0.1"},
					},
				},
			},
			nics: []*apiv2.MachineNic{
				{
					Mac: "00:03:00:11:11:01",
				},
				{
					Mac: "00:03:00:11:12:01",
				},
			},
			wantFilePaths: []string{
				"machine/10-lan0.link",
				"machine/10-lan0.network",
				"machine/11-lan1.link",
				"machine/11-lan1.network",
			},
			wantErr: nil,
		},
		{
			name: "render firewall",
			allocation: &apiv2.MachineAllocation{
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Networks: []*apiv2.MachineNetwork{
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Ips:         []string{"10.0.17.2"},
						Vrf:         3981,
					},
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Ips:         []string{"10.0.18.1"},
						Vrf:         3982,
					},
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:         []string{"185.1.2.3"},
						Vrf:         104009,
					},
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:         []string{"100.127.129.1"},
						Vrf:         104010,
					},
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:         []string{"2001::4"},
					},
					{
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Ips:         []string{"10.1.0.1"},
					},
				},
			},
			nics: []*apiv2.MachineNic{
				{
					Mac: "00:03:00:11:11:01",
				},
				{
					Mac: "00:03:00:11:12:01",
				},
			},
			wantFilePaths: []string{
				"firewall/10-lan0.link",
				"firewall/10-lan0.network",
				"firewall/11-lan1.link",
				"firewall/11-lan1.network",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}

			gotErr := configureLanInterfaces(t.Context(), &Config{
				Log:     slog.Default(),
				fs:      fs,
				Network: network.New(tt.allocation),
				Nics:    tt.nics,
			})

			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			for _, name := range tt.wantFilePaths {
				content, err := fs.ReadFile(path.Join(systemdNetworkPath, path.Base(name)))
				require.NoError(t, err)

				if diff := cmp.Diff(mustReadExpected(name), string(content)); diff != "" {
					t.Errorf("diff (+got -want):\n%s", diff)
				}
			}
		})
	}
}

func mustReadExpected(name string) string {
	tpl, err := expectedInterfaceFiles.ReadFile(path.Join("test", name))
	if err != nil {
		panic(err)
	}

	return string(tpl)
}
