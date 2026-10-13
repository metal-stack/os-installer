# {{ .Comment }}
[Match]
Name=lo

[Address]
Address=127.0.0.1/8
{{- range .CIDRs }}

[Address]
Address={{ . }}
{{- end }}
