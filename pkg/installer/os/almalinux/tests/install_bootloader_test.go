package almalinux_test

import (
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "github.com/metal-stack/os-installer/api/v1"
	"github.com/metal-stack/os-installer/pkg/exec"
	"github.com/metal-stack/os-installer/pkg/installer/os/almalinux"
	oscommon "github.com/metal-stack/os-installer/pkg/installer/os/common"
	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sampleMdadmScanOutput = `ARRAY /dev/md/0  metadata=1.0 UUID=42d10089:ee1e0399:445e7550:62b63ec8 name=any:0
ARRAY /dev/md/1  metadata=1.0 UUID=543eb7f8:98d4d986:e669824d:bebe69e5 name=any:1
ARRAY /dev/md/2  metadata=1.0 UUID=fc32a6f0:ee40d9db:87c8c9f3:a8400c8b name=any:2`

	sampleBlkidOutput = `/dev/sda1: UUID="42d10089-ee1e-0399-445e-755062b63ec8" UUID_SUB="cc57c456-0b2f-6345-c597-d861cc6dd8ac" LABEL="any:0" TYPE="linux_raid_member" PARTLABEL="efi" PARTUUID="273985c8-d097-4123-bcd0-80b4e4e14728"
/dev/sda2: UUID="543eb7f8-98d4-d986-e669-824dbebe69e5" UUID_SUB="54748c60-b566-f391-142c-fb78bb1fc6a9" LABEL="any:1" TYPE="linux_raid_member" PARTLABEL="root" PARTUUID="d7863f4e-af7c-47fc-8c03-6ecdc69bc72d"
/dev/sda3: UUID="fc32a6f0-ee40-d9db-87c8-c9f3a8400c8b" UUID_SUB="582e9b4f-f191-e01e-85fd-2f7d969fbef6" LABEL="any:2" TYPE="linux_raid_member" PARTLABEL="varlib" PARTUUID="e8b44f09-b7f7-4e0d-a7c3-d909617d1f05"
/dev/sdb1: UUID="42d10089-ee1e-0399-445e-755062b63ec8" UUID_SUB="61bd5d8b-1bb8-673b-9e61-8c28dccc3812" LABEL="any:0" TYPE="linux_raid_member" PARTLABEL="efi" PARTUUID="13a4c568-57b0-4259-9927-9ac023aaa5f0"
/dev/sdb2: UUID="543eb7f8-98d4-d986-e669-824dbebe69e5" UUID_SUB="e7d01e93-9340-5b90-68f8-d8f815595132" LABEL="any:1" TYPE="linux_raid_member" PARTLABEL="root" PARTUUID="ab11cd86-37b8-4bae-81e5-21fe0a9c9ae0"
/dev/sdb3: UUID="fc32a6f0-ee40-d9db-87c8-c9f3a8400c8b" UUID_SUB="764217ad-1591-a83a-c799-23397f968729" LABEL="any:2" TYPE="linux_raid_member" PARTLABEL="varlib" PARTUUID="9afbf9c1-b2ba-4b46-8db1-e802d26c93b6"
/dev/md1: LABEL="root" UUID="ace079b5-06be-4429-bbf0-081ea4d7d0d9" TYPE="ext4"
/dev/md0: LABEL="efi" UUID="C236-297F" TYPE="vfat"
/dev/md2: LABEL="varlib" UUID="385e8e8e-dbfd-481e-93a4-cba7f4d5fa02" TYPE="ext4"`
)

func Test_os_GrubInstall(t *testing.T) {
	tests := []struct {
		name      string
		cmdLine   string
		details   *v1.MachineDetails
		fsMocks   func(fs *afero.Afero)
		execMocks []test.FakeExecParams
		want      string
		wantErr   error
	}{
		{
			name:    "without raid",
			cmdLine: "console=ttyS1,115200n8 root=UUID=ace079b5-06be-4429-bbf0-081ea4d7d0d9 init=/sbin/init net.ifnames=0 biosdevname=0 nvme_core.io_timeout=300",
			details: &v1.MachineDetails{
				Console:  "ttyS1,115200n8",
				RootUUID: "78cd4dfe-8825-4f45-816e-d284adb0261e",
			},
			execMocks: []test.FakeExecParams{
				{
					WantCmd:  []string{"grub2-mkconfig", "-o", almalinux.GrubConfigPath},
					Output:   "",
					ExitCode: 0,
				},
			},
			want: `GRUB_DEFAULT=0
GRUB_TIMEOUT=5
GRUB_DISTRIBUTOR=almalinux
GRUB_CMDLINE_LINUX_DEFAULT=""
GRUB_CMDLINE_LINUX="console=ttyS1,115200n8 root=UUID=ace079b5-06be-4429-bbf0-081ea4d7d0d9 init=/sbin/init net.ifnames=0 biosdevname=0 nvme_core.io_timeout=300"
GRUB_TERMINAL=serial
GRUB_SERIAL_COMMAND="serial --speed=115200 --unit=1 --word=8"
GRUB_DEVICE=UUID=78cd4dfe-8825-4f45-816e-d284adb0261e
GRUB_ENABLE_BLSCFG=false
`,
		},
		{
			name:    "with raid",
			cmdLine: "console=ttyS1,115200n8 root=UUID=ace079b5-06be-4429-bbf0-081ea4d7d0d9 init=/sbin/init net.ifnames=0 biosdevname=0 nvme_core.io_timeout=300",
			details: &v1.MachineDetails{
				RaidEnabled: true,
				RootUUID:    "ace079b5-06be-4429-bbf0-081ea4d7d0d9",
				Console:     "ttyS1,115200n8",
			},
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, fs.WriteFile("/boot/System.map-5.14.0-503.19.1.el9_5.x86_64", []byte{}, 0600))
				require.NoError(t, fs.WriteFile("/boot/vmlinuz-5.14.0-503.19.1.el9_5.x86_64", []byte{}, 0755))
				require.NoError(t, fs.WriteFile("/boot/initramfs-5.14.0-503.19.1.el9_5.x86_64.img", []byte{}, 0600))
			},
			execMocks: []test.FakeExecParams{
				{
					WantCmd:  []string{"grub2-mkconfig", "-o", almalinux.GrubConfigPath},
					Output:   "",
					ExitCode: 0,
				},
				{
					WantCmd:  []string{"mdadm", "--examine", "--scan"},
					Output:   sampleMdadmScanOutput,
					ExitCode: 0,
				},
				{
					WantCmd:  []string{"blkid"},
					Output:   sampleBlkidOutput,
					ExitCode: 0,
				},
				{
					WantCmd:  []string{"efibootmgr", "-c", "-d", "/dev/sda1", "-p1", "-l", "\\\\EFI\\\\almalinux\\\\shimx64.efi", "-L", "almalinux"},
					Output:   "",
					ExitCode: 0,
				},
				{
					WantCmd:  []string{"efibootmgr", "-c", "-d", "/dev/sdb1", "-p1", "-l", "\\\\EFI\\\\almalinux\\\\shimx64.efi", "-L", "almalinux"},
					Output:   "",
					ExitCode: 0,
				},
				{
					WantCmd: []string{
						"dracut",
						"--mdadmconf",
						"--kver", "5.14.0-503.19.1.el9_5.x86_64",
						"--kmoddir", "/lib/modules/5.14.0-503.19.1.el9_5.x86_64",
						"--include", "/lib/modules/5.14.0-503.19.1.el9_5.x86_64", "/lib/modules/5.14.0-503.19.1.el9_5.x86_64",
						"--fstab",
						"--add=dm mdraid",
						"--add-drivers=raid0 raid1",
						"--hostonly",
						"--force",
					},
					Output:   "",
					ExitCode: 0,
				},
			},
			want: `GRUB_DEFAULT=0
GRUB_TIMEOUT=5
GRUB_DISTRIBUTOR=almalinux
GRUB_CMDLINE_LINUX_DEFAULT=""
GRUB_CMDLINE_LINUX="console=ttyS1,115200n8 root=UUID=ace079b5-06be-4429-bbf0-081ea4d7d0d9 init=/sbin/init net.ifnames=0 biosdevname=0 nvme_core.io_timeout=300"
GRUB_TERMINAL=serial
GRUB_SERIAL_COMMAND="serial --speed=115200 --unit=1 --word=8"
GRUB_DEVICE=UUID=ace079b5-06be-4429-bbf0-081ea4d7d0d9
GRUB_ENABLE_BLSCFG=false
`,
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

			d := almalinux.New(&oscommon.Config{
				Log:            log,
				Fs:             fs,
				Exec:           exec.New(log).WithCommandFn(test.FakeCmd(t, tt.execMocks...)),
				MachineDetails: tt.details,
			})

			gotErr := d.GrubInstall(t.Context(), tt.cmdLine)
			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			content, err := fs.ReadFile(almalinux.DefaultGrubPath)
			require.NoError(t, err)

			assert.Equal(t, tt.want, string(content))
		})
	}
}
