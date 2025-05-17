package talkops

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/r3labs/sse/v2"
	"reflect"
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
	config := s.useConfig()
	callbacks, _ := config["callbacks"].(map[string]reflect.Value)
	functions, _ := config["functions"].(map[string]reflect.Value)
	parameters, _ := config["parameters"].([]*Parameter)
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
		functionArgs, _ := event["defaultArgs"].(map[string]interface{})
		for k, v := range args {
			functionArgs[k] = v
		}
		if function, ok := functions[name]; ok {
			go func() {
				result := function.Call([]reflect.Value{reflect.ValueOf(functionArgs)})
				if len(result) > 0 {
					event["output"] = result[0].Interface()
				}
				publisher.PublishEvent(event)
			}()
		}
		return
	case "boot":
		params, _ := event["parameters"].(map[string]interface{})
		for name, value := range params {
			for i := range parameters {
				if parameters[i].Name == name {
					fmt.Println(parameters[i].Name)
					fmt.Println(value)
					if value == nil {
						parameters[i].SetValue("")
					} else {
						parameters[i].SetValue(fmt.Sprintf("%v", value))
					}
				}
			}
		}
		ready := true
		for i := range parameters {
			if !parameters[i].Optional && !parameters[i].HasValue() {
				ready = false
			}
		}
		publisher.PublishState()
		if !ready {
			return
		}
	}
	for _, t := range EventTypes {
		if eventType == t {
			if callback, ok := callbacks[eventType]; ok {
				args, _ := event["args"].(map[string]interface{})
				go func() {
					callback.Call([]reflect.Value{reflect.ValueOf(args)})
				}()
			}
			return
		}
	}
}
