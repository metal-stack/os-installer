package installer

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func Test_installer_validateRunningInEfiMode(t *testing.T) {
	tests := []struct {
		name    string
		fsMocks func(fs *afero.Afero)
		wantErr error
	}{
		{
			name: "is efi",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, fs.WriteFile("/sys/firmware/efi", []byte(""), 0755))
				require.NoError(t, fs.WriteFile("/sys/class/dmi", []byte(""), 0755))
			},
			wantErr: nil,
		},
		{
			name: "is not efi",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, fs.WriteFile("/sys/class/dmi", []byte(""), 0755))
			},
			wantErr: fmt.Errorf("not running efi mode"),
		},
		{
			name:    "is not efi but virtual",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &installer{
				log: slog.Default(),
				fs: &afero.Afero{
					Fs: afero.NewMemMapFs(),
				},
			}

			if tt.fsMocks != nil {
				tt.fsMocks(i.fs)
			}

			err := i.validateRunningInEfiMode(t.Context())
			if diff := cmp.Diff(tt.wantErr, err, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n %s", diff)
			}
		})
	}
}
