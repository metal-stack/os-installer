package tailscale

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
	//go:embed test/tailscale.service
	expectedTailscaleSystemdUnit string
	//go:embed test/tailscaled.service
	expectedTailscaledSystemdUnit string
)

func TestWriteSystemdUnit(t *testing.T) {
	tests := []struct {
		name                  string
		c                     *TemplateData
		wantTailscaleService  string
		wantTailscaledService string
		wantChanged           bool
		wantErr               error
	}{
		{
			name: "render",
			c: &TemplateData{
				Comment:         `Do not edit.`,
				DefaultRouteVrf: "vrf104009",
				TailscaledPort:  "41161",
				MachineID:       "c0115b51-5e4d-4f92-85c8-1cc504eafdd2",
				AuthKey:         "a-authkey",
				Address:         "headscale.metal-stack.io",
			},
			wantTailscaleService:  expectedTailscaleSystemdUnit,
			wantTailscaledService: expectedTailscaledSystemdUnit,
			wantChanged:           true,
			wantErr:               nil,
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

			content, err := fs.ReadFile(tailscaleServiceUnitPath)
			require.NoError(t, err)

			if diff := cmp.Diff(tt.wantTailscaleService, string(content)); diff != "" {
				t.Errorf("diff (+got -want):\n%s", diff)
			}

			content, err = fs.ReadFile(tailscaledServiceUnitPath)
			require.NoError(t, err)

			if diff := cmp.Diff(tt.wantTailscaledService, string(content)); diff != "" {
				t.Errorf("diff (+got -want):\n%s", diff)
			}
		})
	}
}
