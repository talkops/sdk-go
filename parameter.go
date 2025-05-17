package talkops

import (
	"fmt"
	"os"
	"regexp"
)

type Parameter struct {
	Name string `json:"name"`
	Description string `json:"description"`
	Value string `json:"-"`
	DefaultValue string `json:"defaultValue"`
	AvailableValues []string `json:"availableValues"`
	PossibleValues []string `json:"possibleValues"`
	Optional bool `json:"optional"`
	Type string `json:"type"`
	Env bool `json:"env"`
}

var AllowedTypes = []string{
	"text",
	"password",
	"textarea",
	"email",
	"search",
	"tel",
	"number",
	"url",
	"time",
	"date",
	"datetime-local",
	"select",
	"color",
}

func NewParameter(name string) *Parameter {
	if name == "" {
		panic("name must be a non-empty string")
	}

	matched, err := regexp.MatchString(`^[A-Z0-9_]+$`, name)
	if err != nil || !matched {
		panic("name must contain only uppercase letters, numbers, and underscores")
	}

	return &Parameter{
		Name: name,
		Description: "",
		Value: "",
		DefaultValue: "",
		AvailableValues: []string{},
		PossibleValues: []string{},
		Optional: false,
		Type: "text",
		Env: os.Getenv(name) != "",
	}
}

func (p *Parameter) SetOptional(optional bool) *Parameter {
	p.Optional = optional
	return p
}

func (p *Parameter) SetDescription(description string) *Parameter {
	if description == "" {
		panic("description must be a non-empty string")
	}
	p.Description = description
	return p
}

func (p *Parameter) SetDefaultValue(defaultValue string) *Parameter {
	p.DefaultValue = defaultValue
	return p
}

func (p *Parameter) SetType(type_ string) *Parameter {
	for _, t := range AllowedTypes {
		if t == type_ {
			p.Type = type_
			return p
		}
	}
	panic(fmt.Sprintf("type must be one of: %v", AllowedTypes))
}

func (p *Parameter) GetValue() string {
	if envValue := os.Getenv(p.Name); envValue != "" {
		return envValue
	}
	if p.Value != "" {
		return p.Value
	}
	return p.DefaultValue
}

func (p *Parameter) SetValue(value string) *Parameter {
	p.Value = value
	return p
}

func (p *Parameter) HasValue() bool {
	return p.GetValue() != ""
}

func (p *Parameter) SetAvailableValues(values []string) *Parameter {
	if len(values) == 0 {
		panic("availableValues must be a non-empty list")
	}
	for _, v := range values {
		if v == "" {
			panic("each value in availableValues must be a non-empty string")
		}
	}
	p.AvailableValues = values
	return p
}

func (p *Parameter) SetPossibleValues(values []string) *Parameter {
	if len(values) == 0 {
		panic("possibleValues must be a non-empty list")
	}
	for _, v := range values {
		if v == "" {
			panic("each value in possibleValues must be a non-empty string")
		}
	}
	p.PossibleValues = values
	return p
}
