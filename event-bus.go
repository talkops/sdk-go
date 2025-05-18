package talkops

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"reflect"
	"sync"
	"time"
)

type EventBus struct {
	useState func() map[string]interface{}
	useConfig func() map[string]interface{}
	lastEventState string
	client net.Conn
	writeMutex sync.Mutex
}

func NewEventBus(useState func() map[string]interface{}, useConfig func() map[string]interface{}) *EventBus {
	conn, _ := net.Dial("unix", os.Getenv("TALKOPS_SOCKET"))
	eb := &EventBus{
		useState: useState,
		useConfig: useConfig,
		client: conn,
	}
	go eb.PublishEvent(map[string]interface{}{
		"type": "init",
	})
	go eb.publishStatePeriodically()
	go eb.listen()
	return eb
}

func (eb *EventBus) PublishState() {
	event := map[string]interface{}{
		"type": "state",
		"state": eb.useState(),
	}
	eventJSON, _ := json.Marshal(event)
	eb.lastEventState = string(eventJSON)
	eb.PublishEvent(event)
}

func (eb *EventBus) PublishEvent(event map[string]interface{}) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.Encode(event)
	eb.writeMutex.Lock()
	defer eb.writeMutex.Unlock()
	eb.client.Write(buf.Bytes())
}

func (eb *EventBus) listen() {
	scanner := bufio.NewScanner(eb.client)
	for scanner.Scan() {
		var event map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &event); err == nil {
			go eb.onEvent(event)
		}
	}
}

func (eb *EventBus) publishStatePeriodically() {
	for {
		time.Sleep(500 * time.Millisecond)
		event := map[string]interface{}{
			"type": "state",
			"state": eb.useState(),
		}
		eventJSON, _ := json.Marshal(event)
		lastEventState := string(eventJSON)
		if eb.lastEventState != lastEventState {
			eb.lastEventState = lastEventState
			go eb.PublishEvent(event)
		}
	}
}

func (eb *EventBus) onEvent(event map[string]interface{}) {
	config := eb.useConfig()
	callbacks, _ := config["callbacks"].(map[string]reflect.Value)
	functions, _ := config["functions"].(map[string]reflect.Value)
	parameters, _ := config["parameters"].([]*Parameter)
	eventType, _ := event["type"].(string)
	switch eventType {
	case "boot":
		params, _ := event["parameters"].(map[string]interface{})
		for name, value := range params {
			for i := range parameters {
				if parameters[i].Name == name {
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
		eb.PublishState()
		if !ready {
			return
		}
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
				eb.PublishEvent(event)
			}()
		}
		return
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
