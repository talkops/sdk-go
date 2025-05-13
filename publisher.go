package talkops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Publisher struct {
	useConfig func() map[string]interface{}
	useState func() map[string]interface{}
	lastEventState string
	lastPingAt int64
}

func NewPublisher(useConfig func() map[string]interface{}, useState func() map[string]interface{}) *Publisher {
	p := &Publisher{
		useConfig: useConfig,
		useState:  useState,
	}
	p.publishData(`{"type":"init"}`)
	go p.publishStatePeriodically()
	return p
}

func (p *Publisher) PublishState() {
	event := map[string]interface{}{
		"type": "state",
		"state": p.useState(),
	}
	eventJSON, _ := json.Marshal(event)
	p.lastEventState = string(eventJSON)
	p.PublishEvent(event)
}

func (p *Publisher) OnPing() {
	p.lastPingAt = time.Now().UnixMilli()
	p.PublishEvent(map[string]interface{}{"type": "pong"})
}

func (p *Publisher) PublishEvent(event map[string]interface{}) {
	if p.lastPingAt > 0 && p.lastPingAt < time.Now().UnixMilli()-6000 {
		return
	}
	eventJSON, _ := json.Marshal(event)
	p.publishData(string(eventJSON))
}

func (p *Publisher) publishData(data string) {
	config := p.useConfig()
	mercure := config["mercure"].(map[string]interface{})
	form := url.Values{}
	form.Set("topic", mercure["publisher"].(map[string]interface{})["topic"].(string))
	form.Set("data", data)
	req, _ := http.NewRequest("POST", mercure["url"].(string), bytes.NewBufferString(form.Encode()))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", mercure["publisher"].(map[string]interface{})["token"].(string)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	client.Do(req)
}

func (p *Publisher) publishStatePeriodically() {
	for {
		time.Sleep(200 * time.Millisecond)
		event := map[string]interface{}{
			"type":  "state",
			"state": p.useState(),
		}
		eventJSON, _ := json.Marshal(event)
		lastEventState := string(eventJSON)
		if p.lastEventState != lastEventState {
			fmt.Println(lastEventState)
			fmt.Println(p.lastEventState)
			p.lastEventState = lastEventState
			p.PublishEvent(event)
		}
	}
}
