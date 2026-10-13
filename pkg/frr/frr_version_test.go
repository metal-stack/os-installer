package frr

import (
	"log/slog"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/stretchr/testify/require"
)

func TestDetectVersion(t *testing.T) {
	tests := []struct {
		name      string
		cmdoutput string
		want      *semver.Version
		wantErr   error
	}{
		{
			name: "frr 10.4",
			cmdoutput: `
		vtysh -h
		Usage : vtysh [OPTION...]
		Integrated shell for FRR (version 10.4.3).
		`,
			want:    semver.MustParse("10.4.3"),
			wantErr: nil,
		},
		{
			name: "frr 8.4",
			cmdoutput: `
		Integrated shell for FRR (version 8.4.4).
		Configured with:
		    '--build=x86_64-linux-gnu' '--prefix=/usr' '--includedir=${prefix}/include' '--mandir=${prefix}/share/man' '--infodir=${prefix}/share/info' '--sysconfdir=/etc' '--localstatedir=/var' '--disable-option-checking' '--disable-silent-rules' '--libdir=${prefix}/lib/x86_64-linux-gnu' '--libexecdir=${prefix}/lib/x86_64-linux-gnu' '--disable-maintainer-mode' '--localstatedir=/var/run/frr' '--sbindir=/usr/lib/frr' '--sysconfdir=/etc/frr' '--with-vtysh-pager=/usr/bin/pager' '--libdir=/usr/lib/x86_64-linux-gnu/frr' '--with-moduledir=/usr/lib/x86_64-linux-gnu/frr/modules' '--disable-dependency-tracking' '--enable-rpki' '--disable-scripting' '--disable-pim6d' '--with-libpam' '--enable-doc' '--enable-doc-html' '--enable-snmp' '--enable-fpm' '--disable-protobuf' '--disable-zeromq' '--enable-ospfapi' '--enable-bgp-vnc' '--enable-multipath=256' '--enable-user=frr' '--enable-group=frr' '--enable-vty-group=frrvty' '--enable-configfile-mask=0640' '--enable-logfile-mask=0640' 'build_alias=x86_64-linux-gnu' 'PYTHON=python3'

		-b, --boot               Execute boot startup configuration
		-c, --command            Execute argument as command
		-d, --daemon             Connect only to the specified daemon
		-f, --inputfile          Execute commands from specific file and exit
		-E, --echo               Echo prompt and command in -c mode
		-C, --dryrun             Check configuration for validity and exit
		-m, --markfile           Mark input file with context end
		    --vty_socket         Override vty socket path
		    --config_dir         Override config directory path
		`,
			want:    semver.MustParse("8.4.4"),
			wantErr: nil,
		},

		{
			name: "10.4.1",
			cmdoutput: `Usage : vtysh [OPTION...]

Integrated shell for FRR (version 10.4.1). 
Configured with:
    '--build=x86_64-linux-gnu' '--prefix=/usr' '--includedir=${prefix}/include' '--mandir=${prefix}/share/man' '--infodir=${prefix}/share/info' '--sysconfdir=/etc' '--localstatedir=/var' '--disable-option-checking' '--disable-silent-rules' '--libdir=${prefix}/lib/x86_64-linux-gnu' '--libexecdir=${prefix}/lib/x86_64-linux-gnu' '--disable-maintainer-mode' '--sbindir=/usr/lib/frr' '--with-vtysh-pager=/usr/bin/pager' '--libdir=/usr/lib/x86_64-linux-gnu/frr' '--with-moduledir=/usr/lib/x86_64-linux-gnu/frr/modules' '--disable-dependency-tracking' '--enable-rpki' '--disable-scripting' '--enable-pim6d' '--disable-grpc' '--with-libpam' '--enable-doc' '--enable-doc-html' '--enable-snmp' '--enable-fpm' '--disable-protobuf' '--disable-zeromq' '--enable-ospfapi' '--enable-bgp-vnc' '--enable-cumulus=yes' '--enable-multipath=256' '--enable-pcre2posix' '--enable-user=frr' '--enable-group=frr' '--enable-vty-group=frrvty' '--enable-configfile-mask=0640' '--enable-logfile-mask=0640' 'build_alias=x86_64-linux-gnu' 'PYTHON=python3'

-b, --boot               Execute boot startup configuration
-c, --command            Execute argument as command
-d, --daemon             Connect only to the specified daemon
-f, --inputfile          Execute commands from specific file and exit
-E, --echo               Echo prompt and command in -c mode
-C, --dryrun             Check configuration for validity and exit
-m, --markfile           Mark input file with context end
    --vty_socket         Override vty socket path
    --config_dir         Override config directory path
-N  --pathspace          Insert prefix into config & socket paths
-u  --user               Run as an unprivileged user
-w, --writeconfig        Write integrated config (frr.conf) and exit
-H, --histfile           Override history file
-t, --timestamp          Print a timestamp before going to shell or reading the configuration
    --no-fork            Don't fork clients to handle daemons (slower for large configs)
    --exec-timeout       Set an idle timeout for this vtysh session
-h, --help               Display this help and exit

Note that multiple commands may be executed from the command
line by passing multiple -c args, or by embedding linefeed
characters in one or more of the commands.`,
			want: semver.MustParse("10.4.1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVersion(slog.Default(), tt.cmdoutput)
			if diff := cmp.Diff(tt.wantErr, err, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}
			if tt.wantErr != nil {
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}
