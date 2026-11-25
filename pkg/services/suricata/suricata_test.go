package suricata

import (
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "embed"
)

var (
	//go:embed test/evebox.service
	expectedEveboxSystemdUnit string
	//go:embed test/suricata-update.service
	expectedSuricataUpdateSystemdUnit string
	//go:embed test/suricata.service
	expectedSuricataSystemdUnit string
	//go:embed test/suricata-ips.service
	expectedSuricataIpsSystemdUnit string
	//go:embed test/suricata.yaml
	expectedSuricataConfig string
	//go:embed test/suricata_defaults
	expectedSuricataDefaults string
)

func TestWriteSystemdUnit(t *testing.T) {
	tests := []struct {
		name                      string
		c                         *TemplateData
		wantSuricataUpdateService string
		wantSuricataService       string
		wantSuricataIpsService    string
		wantSuricataConfig        string
		wantSuricataDefaults      string
		wantEveboxService         string
		wantChanged               bool
		wantErr                   error
	}{
		{
			name: "render",
			c: &TemplateData{
				DefaultRouteVrf: "vrf104009",
				Interface:       "vlan104009",
				TenantVrf:       "vrf70",
				HomeNetPrefixes: []string{"185.1.2.0/24", "185.27.0.0/22", "100.127.129.0/24"},
			},
			wantEveboxService:         expectedEveboxSystemdUnit,
			wantSuricataUpdateService: expectedSuricataUpdateSystemdUnit,
			wantSuricataService:       expectedSuricataSystemdUnit,
			wantSuricataIpsService:    expectedSuricataIpsSystemdUnit,
			wantSuricataConfig:        expectedSuricataConfig,
			wantSuricataDefaults:      expectedSuricataDefaults,
			wantChanged:               true,
			wantErr:                   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}

			gotChanged, gotErr := WriteSystemdUnit(t.Context(), &Config{
				Log:    slog.Default(),
				Reload: false,
				fs:     fs,
			}, tt.c)

			assert.Equal(t, tt.wantChanged, gotChanged)

			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			content, err := fs.ReadFile(eveboxServiceUnitPath)
			require.NoError(t, err)

			if diff := cmp.Diff(tt.wantEveboxService, string(content)); diff != "" {
				t.Errorf("diff (+got -want):\n%s", diff)
			}

			content, err = fs.ReadFile(suricataUpdateServiceUnitPath)
			require.NoError(t, err)

			if diff := cmp.Diff(tt.wantSuricataUpdateService, string(content)); diff != "" {
				t.Errorf("diff (+got -want):\n%s", diff)
			}

			content, err = fs.ReadFile(suricataServiceUnitPath)
			require.NoError(t, err)

			if diff := cmp.Diff(tt.wantSuricataService, string(content)); diff != "" {
				t.Errorf("diff (+got -want):\n%s", diff)
			}

			content, err = fs.ReadFile(suricataIpsServicePath)
			require.NoError(t, err)

			if diff := cmp.Diff(tt.wantSuricataIpsService, string(content)); diff != "" {
				t.Errorf("diff (+got -want):\n%s", diff)
			}

			content, err = fs.ReadFile(suricataConfigPath)
			require.NoError(t, err)

			if diff := cmp.Diff(tt.wantSuricataConfig, string(content)); diff != "" {
				t.Errorf("diff (+got -want):\n%s", diff)
			}

			content, err = fs.ReadFile(suricataDefaultsPath)
			require.NoError(t, err)

			if diff := cmp.Diff(tt.wantSuricataDefaults, string(content)); diff != "" {
				t.Errorf("diff (+got -want):\n%s", diff)
			}
		})
	}
}

func TestHomeNetPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		prefixes []string
		want     []string
		wantErr  string
	}{
		{
			name:     "no prefixes",
			prefixes: nil,
			want:     nil,
		},
		{
			name: "private prefixes are dropped, public ones kept and deduplicated",
			prefixes: []string{
				"10.0.16.0/22",
				"10.0.12.0/22",
				"172.16.5.0/24",
				"192.168.1.0/24",
				"185.1.2.0/24",
				"185.27.0.0/22",
				"185.1.2.0/24",
				"100.127.129.0/24",
				"2001:db8::/48",
			},
			want: []string{"185.1.2.0/24", "185.27.0.0/22", "100.127.129.0/24", "2001:db8::/48"},
		},
		{
			name:     "prefix larger than a private range is kept",
			prefixes: []string{"10.0.0.0/7"},
			want:     []string{"10.0.0.0/7"},
		},
		{
			name:     "invalid prefix",
			prefixes: []string{"foo"},
			wantErr:  `unable to parse prefix "foo": netip.ParsePrefix("foo"): no '/'`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HomeNetPrefixes(tt.prefixes)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
