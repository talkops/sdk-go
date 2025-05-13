package talkops

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	//"github.com/openai/openai-go"
	"os"
	"runtime/debug"
	"strings"
)

type Function struct {
	Name string `json:"name"`
	Description string `json:"description"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Handler func(map[string]interface{}) string `json:"-"`
}

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
	if err := json.Unmarshal(categoriesData, &Categories); err != nil {
		panic(fmt.Sprintf("Failed to parse categories.json: %v", err))
	}
	if err := json.Unmarshal(eventTypesData, &EventTypes); err != nil {
		panic(fmt.Sprintf("Failed to parse event-types.json: %v", err))
	}
}

func GetSdkVersion() string {
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
	callbacks         map[string]func(args ...interface{})
	category          string
	demo              bool
	features          []string
	functions         []interface{}
	functionSchemas   []map[string]interface{}
	icon              string
	installationSteps []string
	instructions      string
	name              string
	parameters        []Parameter
	publisher         *Publisher
	softwareVersion   string
	started           bool
	token             string
	website           string
}

func NewExtension() *Extension {
	return &Extension{
		callbacks: make(map[string]func(args ...interface{})),
		token: os.Getenv("TALKOPS_TOKEN"),
	}
}

func NewExtensionFromToken(token string) *Extension {
	return &Extension{
		callbacks: make(map[string]func(args ...interface{})),
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

func (e *Extension) SetParameters(params []Parameter) *Extension {
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

func (e *Extension) SetFunctions(functions []interface{}) *Extension {
	e.functions = functions
	return e
}

func (e *Extension) On(eventType string, cb func(args ...interface{})) *Extension {
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
	e.callbacks[eventType] = cb
	return e
}

func (e *Extension) Start() *Extension {
	if e.started {
		return e
	}
	e.started = true
	mercure := parseToken(e.token)
	version := GetSdkVersion()
	fmt.Println(os.Getenv("ENV"))
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

func parseToken(token string) map[string]interface{} {
	if token == "" {
		return map[string]interface{}{}
	}
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		panic(err)
	}
	var result map[string]interface{}
	err = json.Unmarshal(decoded, &result)
	if err != nil {
		panic(err)
	}
	return result
}

func (e *Extension) EnableAlarm() *Extension {
	if e.publisher != nil {
		e.publisher.PublishEvent(map[string]interface{}{"type": "alarm"})
	}
	return e
}

func (e *Extension) SendMedias(medias ...interface{}) *Extension {
	if e.publisher != nil {
		var mediaList []interface{}
		for _, m := range medias {
			mediaList = append(mediaList, m)
		}
		e.publisher.PublishEvent(map[string]interface{}{
			"type":   "medias",
			"medias": mediaList,
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
