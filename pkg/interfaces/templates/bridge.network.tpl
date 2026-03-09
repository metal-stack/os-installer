# {{ .Comment }}
[Match]
Name=bridge

[Network]
{{- range .EVPNIfaces }}
VLAN=vlan{{ .VrfID }}
{{- end }}
{{- range .EVPNIfaces }}

[BridgeVLAN]
VLAN={{ .VlanID }}
{{- end }}
