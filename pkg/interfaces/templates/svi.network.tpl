# {{ .Comment }}
[Match]
Name=vlan{{ .EVPNIface.VrfID }}

[Link]
MTUBytes=9000

[Network]
VRF=vrf{{ .EVPNIface.VrfID }}
{{- range .EVPNIface.CIDRs }}
Address={{ . }}
{{- end }}
