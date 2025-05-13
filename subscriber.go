package talkops

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/r3labs/sse/v2"
)

type Subscriber struct {
	useConfig func() map[string]interface{}
}

func NewSubscriber(useConfig func() map[string]interface{}) *Subscriber {
	config := useConfig()
	mercure := config["mercure"].(map[string]interface{})
	topic := mercure["subscriber"].(map[string]interface{})["topic"].(string)
	url := mercure["url"].(string)
	token := mercure["subscriber"].(map[string]interface{})["token"].(string)

	client := sse.NewClient(fmt.Sprintf("%s?topic=%s", url, topic))
	if token != "" {
		client.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	}

	s := &Subscriber{
		useConfig: useConfig,
	}

	client.SubscribeRaw(func(msg *sse.Event) {
		var event map[string]interface{}
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			return
		}
		s.onEvent(event)
	})

	return s
}

func (s *Subscriber) onEvent(event map[string]interface{}) {
	fmt.Println(event)
	config := s.useConfig()
	callbacks, _ := config["callbacks"].(map[string]func(args ...interface{}))
	functions, _ := config["functions"].([]func(args ...interface{}) interface{})
	parameters, _ := config["parameters"].([]Parameter)
	publisher, _ := config["publisher"].(*Publisher)
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
	for _, t := range EventTypes {
		if eventType == t {
			if cb, ok := callbacks[eventType]; ok {
				cb(event)
			}
			return
		}
	}
}
