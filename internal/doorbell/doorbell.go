package doorbell

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"encoding/json"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Unit struct {
	ID      string
	Name    string
	Address string
}

type BellPress struct {
	UnitID string
	Action string
}

func (c *Controller) subscribe() (<-chan BellPress, error) {

	client := mqtt.NewClient(c.mqttOpts)
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
	log.Printf("Attempting to ring to %s", bellPress.UnitID)
	unit, ok := c.LookupUnit(bellPress.UnitID)
	if !ok {
		log.Printf("No configuration for %s unit", bellPress.UnitID)
		return
	}
	topic := fmt.Sprintf("%s-%s", bellPress.UnitID, c.ntfyTopicSuffix)
	url := fmt.Sprintf("https://ntfy.sh/%s", topic)
	msg := fmt.Sprintf("Ring %s!", unit.Name)
	_, err := http.Post(url, "text/plain", strings.NewReader(msg))
	if err != nil {
		log.Printf("Failed to ring %s: %v", bellPress.UnitID, err)
		return
	}

	log.Printf("Send a ring to %s", bellPress.UnitID)
}
