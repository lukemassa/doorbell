package doorbell

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"encoding/json"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type UnitConfiguration struct{}

type BellPress struct {
	Unit   string
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
		unit := strings.TrimPrefix(msg.Topic(), "zigbee2mqtt/")
		if _, ok := c.units[unit]; !ok {
			return
		}
		var bellPress BellPress

		err := json.Unmarshal(msg.Payload(), &bellPress)
		if err != nil {
			log.Printf("Parsing message: %v", err)
			return
		}
		if bellPress.Action == "" {
			return
		}
		bellPress.Unit = unit
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

	topic := fmt.Sprintf("%s-%s", bellPress.Unit, c.ntfyTopicSuffix)
	url := fmt.Sprintf("https://ntfy.sh/%s", topic)
	msg := fmt.Sprintf("Ring %s!", bellPress.Unit)
	_, err := http.Post(url, "text/plain", strings.NewReader(msg))
	if err != nil {
		log.Printf("Failed to ring %s: %v", bellPress.Unit, err)
	}
}
