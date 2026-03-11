package ubuntu_test

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "github.com/metal-stack/os-installer/api/v1"
	"github.com/metal-stack/os-installer/pkg/exec"
	oscommon "github.com/metal-stack/os-installer/pkg/installer/os/common"
	"github.com/metal-stack/os-installer/pkg/installer/os/ubuntu"
	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/assert/yaml"
	"github.com/stretchr/testify/require"
)

func Test_os_WriteBootInfo(t *testing.T) {
	tests := []struct {
		name    string
		cmdLine string
		fsMocks func(fs *afero.Afero)
		want    *v1.Bootinfo
		wantErr error
	}{
		{
			name:    "boot-info ubuntu",
			cmdLine: "a-cmd-line",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, fs.WriteFile("/boot/System.map-1.2.3", nil, 0700))
				require.NoError(t, fs.WriteFile("/boot/vmlinuz-1.2.3", nil, 0700))
				require.NoError(t, fs.WriteFile("/boot/initrd.img-1.2.3", nil, 0700))
			},
			want: &v1.Bootinfo{
				Initrd:       "/boot/initrd.img-1.2.3",
				Cmdline:      "a-cmd-line",
				Kernel:       "/boot/vmlinuz-1.2.3",
				BootloaderID: "metal-ubuntu",
			},
		},
		{
			name:    "more than one system.map present",
			cmdLine: "a-cmd-line",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, fs.WriteFile("/boot/System.map-1.2.3", nil, 0700))
				require.NoError(t, fs.WriteFile("/boot/System.map-1.2.4", nil, 0700))
				require.NoError(t, fs.WriteFile("/boot/vmlinuz-1.2.3", nil, 0700))
				require.NoError(t, fs.WriteFile("/boot/initrd.img-1.2.3", nil, 0700))
			},
			want:    nil,
			wantErr: fmt.Errorf("more or less than a single System.map found ([/boot/System.map-1.2.3 /boot/System.map-1.2.4]), probably no kernel or more than one kernel installed"),
		},
		{
			name:    "no system.map present",
			cmdLine: "a-cmd-line",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, fs.WriteFile("/boot/vmlinuz-1.2.3", nil, 0700))
				require.NoError(t, fs.WriteFile("/boot/initrd.img-1.2.3", nil, 0700))
			},
			want:    nil,
			wantErr: fmt.Errorf("more or less than a single System.map found ([]), probably no kernel or more than one kernel installed"),
		},
		{
			name:    "no vmlinuz present",
			cmdLine: "a-cmd-line",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, fs.WriteFile("/boot/System.map-1.2.3", nil, 0700))
				require.NoError(t, fs.WriteFile("/boot/initrd.img-1.2.3", nil, 0700))
			},
			want:    nil,
			wantErr: fmt.Errorf("kernel image \"/boot/vmlinuz-1.2.3\" not found"),
		},
		{
			name:    "no ramdisk present",
			cmdLine: "a-cmd-line",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, fs.WriteFile("/boot/System.map-1.2.3", nil, 0700))
				require.NoError(t, fs.WriteFile("/boot/vmlinuz-1.2.3", nil, 0700))
			},
			want:    nil,
			wantErr: fmt.Errorf("ramdisk \"/boot/initrd.img-1.2.3\" not found"),
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
				Log:  log,
				Fs:   fs,
				Exec: exec.New(log).WithCommandFn(test.FakeCmd(t)),
			})

			gotErr := d.WriteBootInfo(t.Context(), tt.cmdLine)
			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			content, err := fs.ReadFile(oscommon.BootInfoPath)
			require.NoError(t, err)

			var bootInfo v1.Bootinfo
			err = yaml.Unmarshal(content, &bootInfo)
			require.NoError(t, err)

			assert.Equal(t, tt.want, &bootInfo)
		})
	}
}
