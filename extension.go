package talkops

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
)

//go:embed readme.tmpl
var ReadmeTemplateData []byte
var ReadmeTemplate string

//go:embed categories.json
var categoriesData []byte
var Categories []string

//go:embed event-types.json
var eventTypesData []byte
var EventTypes []string

func init() {
	ReadmeTemplate = string(ReadmeTemplateData)
	json.Unmarshal(categoriesData, &Categories)
	json.Unmarshal(eventTypesData, &EventTypes)
}

func getSdkVersion() string {
	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/talkops/sdk-go" {
				return strings.TrimPrefix(dep.Version, "v")
			}
		}
	}
	return ""
}

type Extension struct {
	callbacks         map[string]reflect.Value
	category          string
	demo              bool
	features          []string
	functions         map[string]reflect.Value
	functionSchemas   []map[string]interface{}
	icon              string
	installationSteps []string
	instructions      string
	name              string
	parameters        []*Parameter
	publisher         *Publisher
	softwareVersion   string
	started           bool
	token             string
	website           string
}

func NewExtension() *Extension {
	return &Extension{
		callbacks: make(map[string]reflect.Value),
		token: os.Getenv("TALKOPS_TOKEN"),
	}
}

func NewExtensionFromToken(token string) *Extension {
	return &Extension{
		callbacks: make(map[string]reflect.Value),
		token: token,
	}
}

func (e *Extension) SetDemo(demo bool) *Extension {
	e.demo = demo
	return e
}

func (e *Extension) SetName(name string) *Extension {
	if name == "" {
		panic("name must be a non-empty string")
	}
	e.name = name
	return e
}

func (e *Extension) SetIcon(icon string) *Extension {
	if icon == "" {
		panic("icon must be a non-empty string")
	}
	e.icon = icon
	return e
}

func (e *Extension) SetWebsite(website string) *Extension {
	if website == "" {
		panic("website must be a non-empty string")
	}
	tryURL := func(url string) bool {
		return true
	}
	if !tryURL(website) {
		panic("website must be a valid URL")
	}
	e.website = website
	return e
}

func (e *Extension) SetSoftwareVersion(version string) *Extension {
	e.softwareVersion = version
	return e
}

func (e *Extension) SetCategory(category string) *Extension {
	found := false
	for _, c := range Categories {
		if c == category {
			found = true
			break
		}
	}
	if !found {
		panic(fmt.Sprintf("category must be one of: %v", Categories))
	}
	e.category = category
	return e
}

func (e *Extension) SetFeatures(features []string) *Extension {
	for _, f := range features {
		if f == "" {
			panic("features must be non-empty strings")
		}
	}
	e.features = features
	return e
}

func (e *Extension) SetInstallationSteps(steps []string) *Extension {
	for _, s := range steps {
		if s == "" {
			panic("installationSteps must be non-empty strings")
		}
	}
	e.installationSteps = steps
	return e
}

func (e *Extension) SetParameters(params []*Parameter) *Extension {
	e.parameters = params
	return e
}

func (e *Extension) SetInstructions(instructions string) *Extension {
	if instructions == "" {
		panic("instructions must be a non-empty string")
	}
	e.instructions = instructions
	return e
}

func (e *Extension) SetFunctionSchemas(schemas []map[string]interface{}) *Extension {
	e.functionSchemas = schemas
	return e
}

func (e *Extension) SetFunctions(functions map[string]func(args map[string]interface{}) string) *Extension {
	e.functions = make(map[string]reflect.Value)
	for name, fn := range functions {
		e.functions[name] = reflect.ValueOf(fn)
	}
	return e
}

func (e *Extension) On(eventType string, cb func(args map[string]interface{})) *Extension {
	found := false
	for _, t := range EventTypes {
		if t == eventType {
			found = true
			break
		}
	}
	if !found {
		panic(fmt.Sprintf("eventType must be one of: %v", EventTypes))
	}
	e.callbacks[eventType] = reflect.ValueOf(cb)
	return e
}

func (e *Extension) Start() *Extension {
	if e.started {
		return e
	}
	e.started = true
	decoded, _ := base64.StdEncoding.DecodeString(e.token)
	var mercure map[string]interface{}
	json.Unmarshal(decoded, &mercure)
	version := getSdkVersion()
	if os.Getenv("ENV") == "development" {
		NewManifest(func() map[string]interface{} {
			return map[string]interface{}{
				"category": e.category,
				"demo": e.demo,
				"features": e.features,
				"icon": e.icon,
				"name": e.name,
				"sdk": map[string]interface{}{
					"name": "go",
					"version": version,
				},
				"softwareVersion": e.softwareVersion,
				"website": e.website,
			}
		})
		NewReadme(func() map[string]interface{} {
			return map[string]interface{}{
				"features": e.features,
				"name": e.name,
			}
		})
	}
	e.publisher = NewPublisher(
		func() map[string]interface{} {
			return map[string]interface{}{
				"mercure": mercure,
			}
		},
		func() map[string]interface{} {
			return map[string]interface{}{
				"category": e.category,
				"demo": e.demo,
				"icon": e.icon,
				"installationSteps": e.installationSteps,
				"instructions": e.instructions,
				"name": e.name,
				"parameters": e.parameters,
				"sdk": map[string]interface{}{
					"name": "go",
					"version": version,
				},
				"softwareVersion": e.softwareVersion,
				"functionSchemas": e.functionSchemas,
			}
		},
	)
	NewSubscriber(func() map[string]interface{} {
		return map[string]interface{}{
			"callbacks": e.callbacks,
			"extension": e,
			"functions": e.functions,
			"mercure": mercure,
			"parameters": e.parameters,
			"publisher": e.publisher,
		}
	})
	return e
}

func (e *Extension) EnableAlarm() *Extension {
	if e.publisher != nil {
		e.publisher.PublishEvent(map[string]interface{}{"type": "alarm"})
	}
	return e
}

func (e *Extension) SendMedias(medias []*Media) *Extension {
	if e.publisher != nil {
		e.publisher.PublishEvent(map[string]interface{}{
			"type": "medias",
			"medias": medias,
		})
	}
	return e
}

func (e *Extension) SendMessage(text string) *Extension {
	if e.publisher != nil {
		e.publisher.PublishEvent(map[string]interface{}{
			"type": "message",
			"text": text,
		})
	}
	return e
}

func (e *Extension) SendNotification(text string) *Extension {
	if e.publisher != nil {
		e.publisher.PublishEvent(map[string]interface{}{
			"type": "notification",
			"text": text,
		})
	}
	return e
}
