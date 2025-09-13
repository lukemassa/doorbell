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

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("connecting to MQTT broker: %v", token.Error())
	}

	ret := make(chan BellPress)
	topic := "zigbee2mqtt/#"
	token := client.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
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
		ret <- bellPress
	})

	token.Wait()

	err := token.Error()

	if err != nil {
		return nil, fmt.Errorf("subscribing to topic: %v", token.Error())
	}

	log.Println("Listening for Zigbee2MQTT messages...")
	return ret, nil
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
