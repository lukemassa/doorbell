package doorbell

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/lukemassa/clilog"
)

var topicRe = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)

type Notifier interface {
	Notify() error
	Validate(showSecrets bool) error
}

type NtfyNotifier struct {
	topic   string
	message string
}

type ChimeNotifier struct {
	mqttURL string
	address string
}

func (r NtfyNotifier) Notify() error {
	_, err := http.Post(r.url(), "text/plain", strings.NewReader(r.message))
	if err != nil {
		return fmt.Errorf("notifying ntfy: %v", err)
	}
	return nil
}

func (r NtfyNotifier) url() string {
	return fmt.Sprintf("https://ntfy.sh/%s", r.topic)
}

func (c ChimeNotifier) Notify() error {

	// This uses a different client than the one that listens, is that what we want?
	opts := mqtt.NewClientOptions().
		AddBroker(c.mqttURL).
		SetClientID("zigbee2mqtt-logger") // TODO NEW CLIENT

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	defer client.Disconnect(250)

	// Publish message
	topic := fmt.Sprintf("zigbee2mqtt/%s/set/alarm", c.address)
	payload := "true"
	token := client.Publish(topic, 0, false, payload)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (n NtfyNotifier) Validate(showSecrets bool) error {
	if showSecrets {
		log.Infof("SECRET ntfy topic is %s", n.topic)
	}
	if !topicRe.MatchString(n.topic) {
		suffix := "<REDACTED>"
		if showSecrets {
			suffix = n.topic
		}
		return fmt.Errorf("ntfy token does not match %s: %s", topicRe, suffix)
	}
	log.Info("ntfy notifier is valid")
	return nil
}

func (c ChimeNotifier) Validate(_ bool) error {
	log.Info("Chime notifier is valid")
	return nil
}
