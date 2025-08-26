package doorbell

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"time"

	"filippo.io/age"
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

	ntfyToken, err := getNtfyToken(config.EncryptedNtfySuffix, config.ageIdentities)
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

func getNtfyToken(ntfyTopicSuffixFile string, identities []age.Identity) (string, error) {

	decoded, err := base64.StdEncoding.DecodeString(ntfyTopicSuffixFile)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %v", err)
	}
	r, err := age.Decrypt(bytes.NewReader(decoded), identities...)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	ntfyTopicSuffixBytes, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	return string(ntfyTopicSuffixBytes), nil
}
