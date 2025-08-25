package doorbell

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const baseHealthURL = "https://hc-ping.com/4003a09f-f033-4f38-82ff-a6a0f010fa50"
const updateFreq = 10 * time.Minute
const maxTemp = 55 // degrees celsius

type BellPress struct {
	Action string
}

type SystemStatus struct {
	temp float64
	err  error
}

type SystemReport struct {
	Temp    float64 `json:"temp"`
	Message string  `json:"message"`
	OK      bool    `json:"ok"`
}

func (s SystemStatus) Report() SystemReport {
	ret := SystemReport{
		Temp: s.temp,
	}
	if s.err != nil {
		ret.Message = fmt.Sprintf("ERROR: %v", s.err)
		ret.OK = false
		return ret
	}
	if s.temp > maxTemp {
		ret.Message = fmt.Sprintf("Temp above threshold %dC", maxTemp)
		ret.OK = false
		return ret
	}
	ret.OK = true
	ret.Message = "All systems normal"
	return ret

}

func updateHealthcheck(status SystemStatus) {
	url := baseHealthURL

	report := status.Report()
	if !report.OK {
		url += "/fail"
	}

	// Encode to JSON
	b, err := json.Marshal(report)
	if err != nil {
		panic(err)
	}

	// Wrap in an io.Reader
	var r io.Reader = bytes.NewReader(b)

	resp, err := http.Post(url, "JSON", r)
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

type Controller struct {
	client mqtt.Client
}

func New(mqttURL string) (*Controller, error) {
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID("zigbee2mqtt-logger")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("connecting to MQTT broker: %v", token.Error())
	}

	return &Controller{
		client: client,
	}, nil

}

func (c *Controller) Run() error {

	err := c.subscribe()
	if err != nil {
		return err
	}

	// TODO: Update system health more quickly if there's an issue, but not too many times
	healthCheckTimer := time.NewTicker(updateFreq)
	systemHealthTimer := time.NewTicker(5 * time.Second)
	defer healthCheckTimer.Stop()
	defer systemHealthTimer.Stop()

	status := c.systemHealth()
	updateHealthcheck(status)
	for {
		select {
		case <-healthCheckTimer.C:
			updateHealthcheck(status)
		case <-systemHealthTimer.C:
			status = c.systemHealth()
		}
	}

}

func (c *Controller) systemHealth() SystemStatus {
	ret := SystemStatus{}
	temp, err := readCPUTemp()
	if err != nil {
		return SystemStatus{
			err: err,
		}
	}
	ret.temp = temp
	return ret
}

func (c *Controller) subscribe() error {

	topic := "zigbee2mqtt/#"
	if token := c.client.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
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
		return fmt.Errorf("subscribing to topic: %v", token.Error())
	}

	fmt.Println("Listening for Zigbee2MQTT messages... (Press Ctrl+C to quit)")
	return nil
}

func getNtfyToken() (string, error) {
	res, err := os.ReadFile("/tmp/notify_topic")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(res)), nil
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

func readCPUTemp() (float64, error) {
	data, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(data))
	milli, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	// Convert millidegrees Celsius → degrees
	return float64(milli) / 1000.0, nil
}
