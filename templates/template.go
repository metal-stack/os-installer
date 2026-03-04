package templates

import (
	"bytes"
	_ "embed"
	"text/template"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
)

type Chrony struct {
	NTPServers []*apiv2.NTPServer
}

//go:embed chrony.conf.tpl
var chronyTemplate string

func RenderChronyTemplate(chronyConfig Chrony) (string, error) {
	templ, err := template.New("chrony").Parse(chronyTemplate)
	if err != nil {
		return "error parsing template", err
	}

	rendered := new(bytes.Buffer)
	err = templ.Execute(rendered, chronyConfig)
	if err != nil {
		return "error writing to template file", err
	}
	return rendered.String(), nil
}
