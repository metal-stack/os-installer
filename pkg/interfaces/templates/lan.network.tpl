# {{ .Comment }}
[Match]
Name=lan{{ .Index }}

[Network]
IPv6AcceptRA=no
{{- range .VxlanIDs }}
VXLAN=vni{{ . }}
{{- end }}
