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
	tmpl, err := template.ParseFiles("/app/readme.tmpl")
	if err != nil {
		return
	}
	f, err := os.Create("/app/README.md")
	if err != nil {
		return
	}
	defer f.Close()
	tmpl.Execute(f, r.getter())
}
