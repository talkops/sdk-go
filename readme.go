package talkops

import (
	"os"
	"text/template"
)

type Readme struct {
	getter func() map[string]interface{}
}

func NewReadme(getter func() map[string]interface{}) *Readme {
	r := &Readme{getter: getter}
	r.generate()
	return r
}

func (r *Readme) generate() {
	tmpl, _ := template.New("readme").Parse(ReadmeTemplate)
	f, _ := os.Create("/app/README.md")
	defer f.Close()
	tmpl.Execute(f, r.getter())
}
