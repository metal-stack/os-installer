{{- /*gotype: github.com/metal-stack/os-installer/pkg/network.IfacesData*/ -}}
{{ .Loopback.Comment }}
[Match]
Name=lo

[Address]
Address=127.0.0.1/8
{{- range .Loopback.IPs }}

[Address]
Address={{ . }}
{{- end }}