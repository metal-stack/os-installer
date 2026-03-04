{{- /*gotype: github.com/metal-stack/os-installer/pkg/network.IfacesData*/ -}}
{{ .Comment }}
[NetDev]
Name=bridge
Kind=bridge
MTUBytes=9000

[Bridge]
DefaultPVID=none
VLANFiltering=yes
