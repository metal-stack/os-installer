package chrony

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
	//go:embed test/default/chrony.conf
	expectedDefaultConfig string
	//go:embed test/custom/chrony.conf
	expectedCustomConfig string
)

func TestWriteSystemdUnit(t *testing.T) {
	tests := []struct {
		name        string
		c           *TemplateData
		wantConfig  string
		wantChanged bool
		wantErr     error
	}{
		{
			name: "render default",
			c: &TemplateData{
				NTPServers: []string{"time.cloudflare.com"},
			},
			wantConfig:  expectedDefaultConfig,
			wantChanged: true,
			wantErr:     nil,
		},
		{
			name: "render custom",
			c: &TemplateData{
				NTPServers: []string{"1.2.3.4", "1.2.3.5"},
			},
			wantConfig:  expectedCustomConfig,
			wantChanged: true,
			wantErr:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}

			gotChanged, gotErr := WriteSystemdUnit(t.Context(), &Config{
				Log:    slog.Default(),
				Reload: false,
				fs:     fs,
			}, tt.c, "vrf104009")

			assert.Equal(t, tt.wantChanged, gotChanged)

			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			content, err := fs.ReadFile(chronyConfigPath)
			require.NoError(t, err)

			if diff := cmp.Diff(tt.wantConfig, string(content)); diff != "" {
				t.Errorf("diff (+got -want):\n%s", diff)
			}
		})
	}
}
