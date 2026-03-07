package nodeexporter

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
	//go:embed test/node-exporter.service
	expectedSystemdUnit string
)

func TestWriteSystemdUnit(t *testing.T) {
	tests := []struct {
		name        string
		c           *TemplateData
		wantService string
		wantChanged bool
		wantErr     error
	}{
		{
			name: "render",
			c: &TemplateData{
				Comment: `Do not edit.`,
			},
			wantService: expectedSystemdUnit,
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
			}, tt.c)

			assert.Equal(t, tt.wantChanged, gotChanged)

			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			content, err := fs.ReadFile(serviceUnitPath)
			require.NoError(t, err)

			if diff := cmp.Diff(tt.wantService, string(content)); diff != "" {
				t.Errorf("diff (+got -want):\n%s", diff)
			}
		})
	}
}
