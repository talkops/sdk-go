package talkops

import (
	"fmt"
	"os"
	"regexp"
)

type Parameter struct {
	Name            string
	Description     string
	Value           string
	DefaultValue    string
	AvailableValues []string
	PossibleValues  []string
	Optional        bool
	Type            string
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

func NewParameter(name string) (*Parameter, error) {
	if name == "" {
		return nil, fmt.Errorf("name must be a non-empty string")
	}

	matched, err := regexp.MatchString(`^[A-Z0-9_]+$`, name)
	if err != nil || !matched {
		return nil, fmt.Errorf("name must contain only uppercase letters, numbers, and underscores")
	}

	return &Parameter{
		Name:     name,
		Type:     "text",
		Optional: false,
	}, nil
}

func (p *Parameter) SetOptional(optional bool) *Parameter {
	p.Optional = optional
	return p
}

func (p *Parameter) SetDescription(description string) (*Parameter, error) {
	if description == "" {
		return nil, fmt.Errorf("description must be a non-empty string")
	}
	p.Description = description
	return p, nil
}

func (p *Parameter) SetDefaultValue(defaultValue string) (*Parameter, error) {
	if defaultValue == "" {
		return nil, fmt.Errorf("defaultValue must be a non-empty string")
	}
	p.DefaultValue = defaultValue
	return p, nil
}

func (p *Parameter) SetType(type_ string) (*Parameter, error) {
	for _, t := range AllowedTypes {
		if t == type_ {
			p.Type = type_
			return p, nil
		}
	}
	return nil, fmt.Errorf("type must be one of: %v", AllowedTypes)
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

func (p *Parameter) SetValue(value string) (*Parameter, error) {
	if value == "" {
		return nil, fmt.Errorf("value must be a non-empty string")
	}
	p.Value = value
	return p, nil
}

func (p *Parameter) HasValue() bool {
	return p.GetValue() != ""
}

func (p *Parameter) SetAvailableValues(values []string) (*Parameter, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("availableValues must be a non-empty list")
	}
	for _, v := range values {
		if v == "" {
			return nil, fmt.Errorf("each value in availableValues must be a non-empty string")
		}
	}
	p.AvailableValues = values
	return p, nil
}

func (p *Parameter) SetPossibleValues(values []string) (*Parameter, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("possibleValues must be a non-empty list")
	}
	for _, v := range values {
		if v == "" {
			return nil, fmt.Errorf("each value in possibleValues must be a non-empty string")
		}
	}
	p.PossibleValues = values
	return p, nil
}

func (p *Parameter) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"name":            p.Name,
		"description":     p.Description,
		"env":             os.Getenv(p.Name) != "",
		"defaultValue":    p.DefaultValue,
		"availableValues": p.AvailableValues,
		"possibleValues":  p.PossibleValues,
		"optional":        p.Optional,
		"type":            p.Type,
	}
}
