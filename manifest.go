package talkops

import (
	"encoding/json"
	"os"
)

type Manifest struct {
	useExtension func() map[string]interface{}
}

func NewManifest(useExtension func() map[string]interface{}) *Manifest {
	m := &Manifest{useExtension: useExtension}
	m.generate()
	return m
}

func (m *Manifest) generate() {
	f, _ := os.Create("/app/manifest.json")
	defer f.Close()
	json.NewEncoder(f).Encode(m.useExtension())
}
