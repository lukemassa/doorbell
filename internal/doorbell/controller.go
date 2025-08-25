package doorbell

import (
	"log"
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
	units           []Unit
}

func NewController(config *Config) (*Controller, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(config.MQTTURL).
		SetClientID("zigbee2mqtt-logger")

	ntfyToken, err := getNtfyToken(config.NtfyTopicSuffixFile)
	if err != nil {
		return nil, err
	}
	var units []Unit
	for unitId, unitConfig := range config.UnitConfigurations {
		units = append(units, Unit{
			ID:      unitId,
			Name:    unitConfig.Name,
			Address: unitConfig.Address,
		})
	}
	return &Controller{
		mqttOpts:        opts,
		ntfyTopicSuffix: ntfyToken,
		units:           units,
	}, nil

}

func (c Controller) LookupUnit(lookup string) (Unit, bool) {
	for _, unit := range c.units {
		if unit.ID == lookup || unit.Address == lookup {
			log.Printf("Found unit %s for lookup %s", unit.Name, lookup)
			return unit, true
		}
	}
	return Unit{}, false
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
