package renderer_test

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	renderer "github.com/metal-stack/os-installer/pkg/template-renderer"
	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_renderer_Render(t *testing.T) {
	tests := []struct {
		name         string
		c            *renderer.Config
		destFile     string
		fsMock       func(fs afero.Afero)
		wantRendered string
		wantChanged  bool
		wantErr      error
	}{
		{
			name: "render an initial unit",
			c: &renderer.Config{
				ServiceName:    "test.service",
				TemplateString: "{{ .Hostname }}",
				Data: map[string]string{
					"Hostname": "foo",
				},
			},
			destFile:     "/hostname",
			wantRendered: "foo",
			wantChanged:  true,
			wantErr:      nil,
		},
		{
			name: "render an initial unit and call validation func",
			c: &renderer.Config{
				ServiceName:    "test.service",
				TemplateString: "{{ .Hostname }}",
				Data: map[string]string{
					"Hostname": "foo",
				},
				Validate: func(path string) error {
					assert.True(t, strings.HasPrefix(path, "/hostname"))
					return fmt.Errorf("a validation error")
				},
			},
			destFile:     "/hostname",
			wantRendered: "",
			wantChanged:  false,
			wantErr:      fmt.Errorf("a validation error"),
		},
		{
			name: "update existing file",
			c: &renderer.Config{
				ServiceName:    "test.service",
				TemplateString: "{{ .Hostname }}",
				Data: map[string]string{
					"Hostname": "foo",
				},
			},
			destFile: "/hostname",
			fsMock: func(fs afero.Afero) {
				require.NoError(t, fs.WriteFile("/hostname", []byte("bar"), os.ModePerm))
			},
			wantRendered: "foo",
			wantChanged:  true,
			wantErr:      nil,
		},
		{
			name: "update existing file that did not change",
			c: &renderer.Config{
				ServiceName:    "test.service",
				TemplateString: "{{ .Hostname }}",
				Data: map[string]string{
					"Hostname": "foo",
				},
			},
			destFile: "/hostname",
			fsMock: func(fs afero.Afero) {
				require.NoError(t, fs.WriteFile("/hostname", []byte("foo"), os.ModePerm))
			},
			wantRendered: "foo",
			wantChanged:  false,
			wantErr:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.Log = slog.Default()
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			tt.c.Fs = fs

			r, err := renderer.New(tt.c)
			require.NoError(t, err)

			if tt.fsMock != nil {
				tt.fsMock(fs)
			}

			gotChanged, gotErr := r.Render(t.Context(), tt.destFile, false) // reload cannot be easily tested here because it interacts with dbus

			assert.Equal(t, tt.wantChanged, gotChanged)

			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			content, err := fs.ReadFile(tt.destFile)
			require.NoError(t, err)

			if diff := cmp.Diff(tt.wantRendered, string(content)); diff != "" {
				t.Errorf("diff (+got -want):\n%s", diff)
			}
		})
	}
}
