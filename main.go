package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"encoding/json"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const baseHealthURL = "https://hc-ping.com/4003a09f-f033-4f38-82ff-a6a0f010fa50"
const updateFreq = 10 * time.Minute

type BellPress struct {
	Action string
}

func getNtfyToken() (string, error) {
	res, err := os.ReadFile("/tmp/notify_topic")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(res)), nil
}

func updateHealthcheck(status bool) {
	url := baseHealthURL
	if !status {
		url += "/fail"
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to post to %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read body from %s: %v", url, err)
		return
	}

	fmt.Printf("Posted to %s: %s\n", url, string(body))
}

func ring(name string) {
	ntfyToken, err := getNtfyToken()
	if err != nil {
		log.Fatal(err)
	}
	topic := fmt.Sprintf("%s-%s", name, ntfyToken)
	url := fmt.Sprintf("https://ntfy.sh/%s", topic)
	msg := fmt.Sprintf("Ring %s!", name)
	_, err = http.Post(url, "text/plain", strings.NewReader(msg))
	if err != nil {
		log.Printf("Failed to ring %s: %v", name, err)
	}
}

func main() {
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID("zigbee2mqtt-logger")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error connecting to MQTT broker: %v", token.Error())
	}

	topic := "zigbee2mqtt/#"
	if token := client.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
		if msg.Topic() != "zigbee2mqtt/first_floor" {
			return
		}
		var bellPress BellPress

		err := json.Unmarshal(msg.Payload(), &bellPress)
		if err != nil {
			log.Printf("Parsing first floor message: %v", err)
			return
		}
		if bellPress.Action == "" {
			return
		}
		fmt.Printf("First floor button was pressed: %s\n", bellPress.Action)
		if bellPress.Action == "single" {
			ring("first_floor")
		}
	}); token.Wait() && token.Error() != nil {
		log.Fatalf("Error subscribing to topic: %v", token.Error())
	}

	fmt.Println("Listening for Zigbee2MQTT messages... (Press Ctrl+C to quit)")

	for {
		updateHealthcheck(true)
		time.Sleep(updateFreq)
	}
}
