package talkops

import (
	"encoding/json"
	"fmt"

	_ "embed"

	"github.com/r3labs/sse/v2"
)

//go:embed event-types.json
var eventTypesData []byte

var eventTypes []string

func init() {
	if err := json.Unmarshal(eventTypesData, &eventTypes); err != nil {
		panic("Impossible de parser event-types.json")
	}
}

type Subscriber struct {
	client    *sse.Client
	useConfig func() map[string]interface{}
}

func NewSubscriber(useConfig func() map[string]interface{}) *Subscriber {
	config := useConfig()
	mercure, ok := config["mercure"].(map[string]interface{})
	if !ok {
		panic("La configuration Mercure est manquante ou invalide")
	}
	topic, ok := mercure["subscriber"].(map[string]interface{})["topic"].(string)
	if !ok {
		panic("Le topic du subscriber est manquant ou invalide")
	}
	url, ok := mercure["url"].(string)
	if !ok {
		panic("L'URL Mercure est manquante ou invalide")
	}
	token, ok := mercure["subscriber"].(map[string]interface{})["token"].(string)
	if !ok {
		token = ""
	}

	client := sse.NewClient(fmt.Sprintf("%s?topic=%s", url, topic))
	if token != "" {
		client.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	}

	s := &Subscriber{
		client:    client,
		useConfig: useConfig,
	}

	go s.listen()
	return s
}

func (s *Subscriber) listen() {
	config := s.useConfig()
	callbacks, _ := config["callbacks"].(map[string]func(args ...interface{}))
	functions, _ := config["functions"].([]func(args ...interface{}) interface{})
	parameters, _ := config["parameters"].([]Parameter)
	publisher, _ := config["publisher"].(*Publisher)
	s.client.SubscribeRaw(func(msg *sse.Event) {
		if len(msg.Data) == 0 {
			return
		}
		var event map[string]interface{}
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			return
		}
		eventType, _ := event["type"].(string)
		switch eventType {
		case "ping":
			if publisher != nil {
				publisher.OnPing()
			}
			return
		case "function_call":
			name, _ := event["name"].(string)
			args, _ := event["args"].(map[string]interface{})
			for _, fn := range functions {
				fnType := fmt.Sprintf("%T", fn)
				if fnType == name {
					// En Go, il n'est pas trivial de récupérer les noms des arguments d'une fonction
					// Il faudrait une convention ou une structure différente pour les fonctions
					// Ici, on suppose que la fonction prend un seul argument map[string]interface{}
					output := fn(args)
					event["output"] = output
					if publisher != nil {
						publisher.PublishEvent(event)
					}
					return
				}
			}
			return
		case "boot":
			params, _ := event["parameters"].(map[string]interface{})
			for name, value := range params {
				for i := range parameters {
					if parameters[i].Name == name {
						parameters[i].SetValue(fmt.Sprintf("%v", value))
					}
				}
			}
			ready := true
			for i := range parameters {
				if !parameters[i].Optional && !parameters[i].HasValue() {
					ready = false
				}
			}
			if publisher != nil {
				publisher.PublishState()
			}
			if !ready {
				return
			}
		}
		// Gestion des event-types standards
		for _, t := range eventTypes {
			if eventType == t {
				if cb, ok := callbacks[eventType]; ok {
					cb(event)
				}
				return
			}
		}
	})
}
