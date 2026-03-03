{{- /*gotype: github.com/metal-stack/os-installer/pkg/network.IfacesData*/ -}}
{{ .Comment }}
[Match]
Name=bridge

[Network]
{{- range .EVPNIfaces }}
VLAN=vlan{{ .VRF.ID }}
{{- end }}
{{- range .EVPNIfaces }}

[BridgeVLAN]
VLAN={{ .SVI.VLANID }}
{{- end }}