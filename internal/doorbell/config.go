package doorbell

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-yaml"

	"filippo.io/age"
)

type UnitConfiguration struct {
	Name    string                  `yaml:"name"`
	Address string                  `yaml:"address"`
	OnPress []NotificationMechanism `yaml:"on_press"`
}

type NotificationMechanism struct {
	NtfySettings  *NtfySettings  `yaml:"ntfy"`
	ChimeSettings *ChimeSettings `yaml:"chime"`
}

type NtfySettings struct {
	EncryptedTopic string `yaml:"encryptedTopic"`
}

type ChimeSettings struct {
	Address string `yaml:"address"`
}

type Config struct {
	MQTTURL            string                       `yaml:"mqttURL"`
	AgeKeyFile         string                       `yaml:"ageKeyFile"`
	UnitConfigurations map[string]UnitConfiguration `yaml:"units"`
}

func NewConfig(content []byte) (*Config, error) {

	var ret Config

	err := yaml.Unmarshal(content, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Config) Controller() (*Controller, error) {
	if c.AgeKeyFile == "" {
		return nil, errors.New("did not set ageKeyFile")
	}

	identities, err := loadIdentities(c.AgeKeyFile)
	if err != nil {
		return nil, fmt.Errorf("loading identities: %v", err)
	}

	var units []Unit
	for unitID, unitConfiguration := range c.UnitConfigurations {
		if len(unitConfiguration.OnPress) == 0 {
			return nil, fmt.Errorf("must set notification mechanism for %s", unitID)
		}
		var notifiers []Notifier

		for i, notificationConfig := range unitConfiguration.OnPress {
			notifier, err := c.getNotifierFromConfig(notificationConfig, unitConfiguration.Name, identities)
			if err != nil {
				return nil, fmt.Errorf("configuring %s notifier for %s: %v", indexToOrdinal(i), unitID, err)
			}
			notifiers = append(notifiers, notifier)

		}

		units = append(units, Unit{
			ID:        unitID,
			Name:      unitConfiguration.Name,
			Address:   unitConfiguration.Address,
			Notifiers: notifiers,
		})
	}
	return &Controller{
		mqttURL: c.MQTTURL,
		units:   units,
	}, nil
}

func loadIdentities(path string) ([]age.Identity, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ids, err := age.ParseIdentities(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func decrypt(content string, identities []age.Identity) (string, error) {

	decoded, err := base64.StdEncoding.DecodeString(content)
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

func (c *Config) getNotifierFromConfig(notificationConfig NotificationMechanism, name string, identities []age.Identity) (Notifier, error) {

	if notificationConfig.NtfySettings != nil {
		topic, err := decrypt(notificationConfig.NtfySettings.EncryptedTopic, identities)
		if err != nil {
			return nil, err
		}
		return &NtfyNotifier{
			topic:   topic,
			message: fmt.Sprintf("Ring for %s", name),
		}, nil
	}
	if notificationConfig.ChimeSettings != nil {
		return &ChimeNotifier{
			address: notificationConfig.ChimeSettings.Address,
			mqttURL: c.MQTTURL, // Same queue
		}, nil
	}

	return nil, errors.New("could not determine which notifier to use")

}

func indexToOrdinal(d int) string {
	number := d + 1
	suffix := "th"
	// TODO: Finish this, deal with like 11th vs 21st
	switch number {
	case 1:
		suffix = "st"
	case 2:
		suffix = "nd"
	case 3:
		suffix = "rd"
	}
	return fmt.Sprintf("%d%s", number, suffix)
}
