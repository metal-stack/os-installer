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
	//go:embed test/suricata-update.service
	expectedSuricataSystemdUnit string
	//go:embed test/suricata.yaml
	expectedSuricataConfig string
	//go:embed test/suricata_defaults
	expectedSuricataDefaults string
)

func TestWriteSystemdUnit(t *testing.T) {
	tests := []struct {
		name                 string
		c                    *TemplateData
		wantSuricataService  string
		wantSuricataConfig   string
		wantSuricataDefaults string
		wantChanged          bool
		wantErr              error
	}{
		{
			name: "render",
			c: &TemplateData{
				DefaultRouteVrf: "vrf104009",
				Interface:       "vlan104009",
			},
			wantSuricataService:  expectedSuricataSystemdUnit,
			wantSuricataConfig:   expectedSuricataConfig,
			wantSuricataDefaults: expectedSuricataDefaults,
			wantChanged:          true,
			wantErr:              nil,
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

			content, err := fs.ReadFile(suricataUpdateServiceUnitPath)
			require.NoError(t, err)

			if diff := cmp.Diff(tt.wantSuricataService, string(content)); diff != "" {
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
