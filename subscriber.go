package talkops

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type Subscriber struct {
	useConfig func() map[string]interface{}
}

func NewSubscriber(useConfig func() map[string]interface{}) *Subscriber {
	s := &Subscriber{
		useConfig: useConfig,
	}
	go s.subscribe()
	return s
}

func (s *Subscriber) subscribe() {
	for {
		config := s.useConfig()
		mercure := config["mercure"].(map[string]interface{})

		url := fmt.Sprintf("%s?topic=%s",
			mercure["url"].(string),
			url.QueryEscape(mercure["subscriber"].(map[string]interface{})["topic"].(string)),
		)

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s",
			mercure["subscriber"].(map[string]interface{})["token"].(string)))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}

			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				var event map[string]interface{}
				if err := json.Unmarshal([]byte(data), &event); err != nil {
					continue
				}
				s.onEvent(event)
			}
		}

		resp.Body.Close()
	}
}

func (s *Subscriber) onEvent(event map[string]interface{}) {
	config := s.useConfig()

	switch event["type"] {
	case "ping":
		config["publisher"].(*Publisher).OnPing()

	case "function_call":
		name := event["name"].(string)
		args := event["args"].(map[string]interface{})
		defaultArgs := event["defaultArgs"].(map[string]interface{})

		for _, fn := range config["functions"].([]func(args ...interface{}) interface{}) {
			if fnName := getFunctionName(fn); fnName == name {
				var fnArgs []interface{}
				for _, argName := range getFunctionArgs(fn) {
					if val, ok := args[argName]; ok {
						fnArgs = append(fnArgs, val)
					} else if val, ok := defaultArgs[argName]; ok {
						fnArgs = append(fnArgs, val)
					}
				}
				output := fn(fnArgs...)
				event["output"] = output
				config["publisher"].(*Publisher).PublishEvent(event)
				return
			}
		}

	case "boot":
		parameters := event["parameters"].(map[string]interface{})
		for name, value := range parameters {
			for _, param := range config["parameters"].([]Parameter) {
				if param.Name == name {
					param.Value = value.(string)
				}
			}
		}

		ready := true
		for _, param := range config["parameters"].([]Parameter) {
			if !param.Optional && param.Value == "" {
				ready = false
				break
			}
		}

		config["publisher"].(*Publisher).PublishState()
		if !ready {
			return
		}

	default:
		if callback, ok := config["callbacks"].(map[string]func(args ...interface{}))[event["type"].(string)]; ok {
			var callbackArgs []interface{}
			if args, ok := event["args"].(map[string]interface{}); ok {
				for _, argName := range getFunctionArgs(callback) {
					if val, ok := args[argName]; ok {
						callbackArgs = append(callbackArgs, val)
					}
				}
			}
			callback(callbackArgs...)
		}
	}
}

func getFunctionName(fn interface{}) string {
	v := reflect.ValueOf(fn)
	t := v.Type()
	if t.Kind() != reflect.Func {
		return ""
	}
	return runtime.FuncForPC(v.Pointer()).Name()
}

func getFunctionArgs(fn interface{}) []string {
	v := reflect.ValueOf(fn)
	t := v.Type()
	if t.Kind() != reflect.Func {
		return nil
	}
	var args []string
	if t.NumIn() == 1 && t.In(0).Kind() == reflect.Slice {
		args = append(args, "args")
	}
	return args
}
