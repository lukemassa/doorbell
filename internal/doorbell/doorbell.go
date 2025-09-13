package doorbell

import (
	"fmt"
	"os"
	"strings"

	log "github.com/lukemassa/clilog"

	"encoding/json"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const mqttClientRole = "doorbell-listener"

type Unit struct {
	ID        string
	Name      string
	Address   string
	Notifiers []Notifier
}

type BellPress struct {
	UnitID string
	Action string
}

type DoorbellClient struct {
	client        mqtt.Client
	bellPressChan chan BellPress
}

func (m *DoorbellClient) Shutdown() {
	m.client.Disconnect(250)
	close(m.bellPressChan)
}

func (c *Controller) subscribe() (DoorbellClient, error) {

	opts := mqtt.NewClientOptions().
		AddBroker(c.mqttURL).
		// add Pid so that when the process restarts, mqtt doesn't get confused about the client identity
		SetClientID(fmt.Sprintf("%s-%d", mqttClientRole, os.Getpid()))

	ret := make(chan BellPress)
	callback := c.createMQTTCallback(ret)
	opts.OnConnect = func(c mqtt.Client) {
		log.Info("Connected to MQTT broker, subscribing...")
		if token := c.Subscribe("zigbee2mqtt/#", 1, callback); token.Wait() && token.Error() != nil {
			log.Errorf("Failed to subscribe: %v", token.Error())
		}
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return DoorbellClient{}, fmt.Errorf("connecting to MQTT broker: %v", token.Error())
	}
	log.Info("Setup Zigbee2MQTT client")
	return DoorbellClient{
		client:        client,
		bellPressChan: ret,
	}, nil
}

// mqttCallback implements the handler for mqtt
func (c *Controller) createMQTTCallback(bellPressChan chan<- BellPress) func(_ mqtt.Client, msg mqtt.Message) {
	return func(_ mqtt.Client, msg mqtt.Message) {
		lookup := strings.TrimPrefix(msg.Topic(), "zigbee2mqtt/")
		unit, ok := c.LookupUnit(lookup)
		if !ok {
			log.Debugf("Zigbee message for unknown topic %s, ignoring", lookup)
			return
		}
		var bellPress BellPress

		err := json.Unmarshal(msg.Payload(), &bellPress)
		if err != nil {
			log.Errorf("Error message for %s: %v", unit.ID, err)
			return
		}
		if bellPress.Action == "" {
			log.Warnf("Message for unit %s did not contain action", unit.ID)
			return
		}
		bellPress.UnitID = unit.ID
		bellPressChan <- bellPress
	}
}

func (c *Controller) Ring(bellPress BellPress) {
	log.Infof("Attempting notifications to %s", bellPress.UnitID)
	unit, ok := c.LookupUnit(bellPress.UnitID)
	if !ok {
		log.Warnf("No configuration for %s unit", bellPress.UnitID)
		return
	}
	for i, notifier := range unit.Notifiers {
		log.Infof("Attempting notifier %d", i+1)
		err := notifier.Notify()
		if err != nil {
			log.Errorf("Error notifying %s: %v", unit.Name, err)
		}
		log.Infof("Success on notifier %d", i+1)
	}
}
