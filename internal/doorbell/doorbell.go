package doorbell

import (
	"fmt"
	"log"
	"os"
	"strings"

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

func (c *Controller) subscribe() (<-chan BellPress, error) {

	opts := mqtt.NewClientOptions().
		AddBroker(c.mqttURL).
		// add Pid so that when the process restarts, mqtt doesn't get confused about the client identity
		SetClientID(fmt.Sprintf("%s-%d", mqttClientRole, os.Getpid()))

	ret := make(chan BellPress)
	callback := c.createMQTTCallback(ret)
	opts.OnConnect = func(c mqtt.Client) {
		log.Println("Connected to MQTT broker, subscribing...")
		if token := c.Subscribe("zigbee2mqtt/#", 1, callback); token.Wait() && token.Error() != nil {
			log.Printf("Failed to subscribe: %v", token.Error())
		}
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("connecting to MQTT broker: %v", token.Error())
	}
	log.Println("Setup Zigbee2MQTT client")
	return ret, nil
}

// mqttCallback implements the handler for mqtt
func (c *Controller) createMQTTCallback(bellPressChan chan<- BellPress) func(_ mqtt.Client, msg mqtt.Message) {
	return func(_ mqtt.Client, msg mqtt.Message) {
		lookup := strings.TrimPrefix(msg.Topic(), "zigbee2mqtt/")
		unit, ok := c.LookupUnit(lookup)
		if !ok {
			log.Printf("Zigbee message for unknown topic %s, ignoring", lookup)
			return
		}
		var bellPress BellPress

		err := json.Unmarshal(msg.Payload(), &bellPress)
		if err != nil {
			log.Printf("Parsing message for %s: %v", unit.ID, err)
			return
		}
		if bellPress.Action == "" {
			log.Printf("Message for unit %s did not contain action", unit.ID)
			return
		}
		bellPress.UnitID = unit.ID
		bellPressChan <- bellPress
	}
}

func (c *Controller) Ring(bellPress BellPress) {
	log.Printf("Attempting notifications to %s", bellPress.UnitID)
	unit, ok := c.LookupUnit(bellPress.UnitID)
	if !ok {
		log.Printf("No configuration for %s unit", bellPress.UnitID)
		return
	}
	for i, notifier := range unit.Notifiers {
		log.Printf("Attempting notifier %d", i+1)
		err := notifier.Notify()
		if err != nil {
			log.Printf("Error notifying %s: %v", unit.Name, err)
		}
		log.Printf("Success on notifier %d", i+1)
	}
}
