{{- /*gotype: github.com/metal-stack/os-installer/pkg/network.IfacesData*/ -}}
{{ .Comment }}
[Match]
Name=lan{{ .Index }}

[Network]
IPv6AcceptRA=no
{{- range .EVPNIfaces }}
VXLAN=vni{{ .VXLAN.ID }}
{{- end }}