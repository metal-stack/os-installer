package renderer_test

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
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
		validationFn func(fs afero.Afero) func(path string) error
		wantRendered string
		wantChanged  bool
		wantErr      error
	}{
		{
			name: "render an initial unit",
			c: &renderer.Config{
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
		{
			name: "verify tmp files look as expected",
			c: &renderer.Config{
				TemplateString: "{{ .Hostname }}",
				Data: map[string]string{
					"Hostname": "foo",
				},
			},
			destFile: "/hostname",
			validationFn: func(fs afero.Afero) func(path string) error {
				return func(path string) error {
					files, err := fs.ReadDir("/")
					require.NoError(t, err)

					require.Len(t, files, 1)

					fileName := files[0].Name()

					assert.True(t, strings.HasPrefix(fileName, "hostname-"))
					fileName = strings.TrimPrefix(fileName, "hostname-")
					_, err = uuid.Parse(fileName)
					require.NoError(t, err, "not a uuid")

					return nil
				}
			},
			wantRendered: "foo",
			wantChanged:  true,
			wantErr:      nil,
		},
		{
			name: "verify tmp files look as expected with complex path",
			c: &renderer.Config{
				TemplateString: "{{ .Rule }}",
				Data: map[string]string{
					"Rule": "allow 1.2.3.4",
				},
				TmpFilePrefix: ".",
			},
			destFile: "/etc/nftables/metal",
			validationFn: func(fs afero.Afero) func(path string) error {
				return func(path string) error {
					files, err := fs.ReadDir("/etc/nftables")
					require.NoError(t, err)

					require.Len(t, files, 1)

					fileName := files[0].Name()

					require.True(t, strings.HasPrefix(fileName, "."))
					fileName = strings.TrimPrefix(fileName, ".")
					assert.True(t, strings.HasPrefix(fileName, "metal-"))
					fileName = strings.TrimPrefix(fileName, "metal-")
					_, err = uuid.Parse(fileName)
					require.NoError(t, err, "not a uuid")

					return nil
				}
			},
			wantRendered: "allow 1.2.3.4",
			wantChanged:  true,
			wantErr:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.Log = slog.Default()
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			tt.c.Fs = fs

			if tt.fsMock != nil {
				tt.fsMock(fs)
			}
			if tt.validationFn != nil {
				tt.c.Validate = tt.validationFn(fs)
			}

			r, err := renderer.New(tt.c)
			require.NoError(t, err)

			gotChanged, gotErr := r.Render(t.Context(), tt.destFile)

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
