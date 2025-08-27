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
	Ntfy *NtfySettings `yaml:"ntfy"`
}
type NtfySettings struct {
	EncryptedTopic string `yaml:"encryptedTopic"`
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

		for _, notificationConfig := range unitConfiguration.OnPress {
			if notificationConfig.Ntfy != nil {
				topic, err := decrypt(notificationConfig.Ntfy.EncryptedTopic, identities)
				if err != nil {
					return nil, err
				}
				notifiers = append(notifiers, &NtfyNotifier{
					topic:   topic,
					message: fmt.Sprintf("Ring for %s", unitConfiguration.Name),
				})
				continue
			}
		}
		if len(notifiers) == 0 {
			return nil, fmt.Errorf("could not configure any notifiers for %s", unitID)
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
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ids, err := age.ParseIdentities(f)
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
