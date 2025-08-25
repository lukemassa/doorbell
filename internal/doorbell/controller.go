package doorbell

import (
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const baseHealthURL = "https://hc-ping.com/4003a09f-f033-4f38-82ff-a6a0f010fa50"
const updateFreq = 10 * time.Minute
const maxTemp = 55 // degrees celsius

type Controller struct {
	mqttOpts        *mqtt.ClientOptions
	ntfyTopicSuffix string
	units           map[string]UnitConfiguration
}

func NewController(config *Config) (*Controller, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(config.MQTTURL).
		SetClientID("zigbee2mqtt-logger")

	ntfyToken, err := getNtfyToken(config.NtfyTopicSuffixFile)
	if err != nil {
		return nil, err
	}
	return &Controller{
		mqttOpts:        opts,
		ntfyTopicSuffix: ntfyToken,
	}, nil

}

func (c *Controller) Run() error {

	ringChan, err := c.subscribe()
	if err != nil {
		return err
	}

	healthCheckTimer := time.NewTicker(updateFreq)
	defer healthCheckTimer.Stop()

	c.updateSystemHealth()
	for {
		select {
		case <-healthCheckTimer.C:
			c.updateSystemHealth()
		case bellPress := <-ringChan:
			c.Ring(bellPress)
		}
	}
}

func getNtfyToken(ntfyTopicSuffixFile string) (string, error) {
	res, err := os.ReadFile(ntfyTopicSuffixFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(res)), nil
}
