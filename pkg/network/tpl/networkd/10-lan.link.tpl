{{- /*gotype: github.com/metal-stack/os-installer/pkg/network.SystemdLinkData*/ -}}
{{ .Comment }}
[Match]
PermanentMACAddress={{ .MAC }}

[Link]
Name=lan{{ .Index }}
NamePolicy=
MTUBytes={{ .MTU }}