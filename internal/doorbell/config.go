package doorbell

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/goccy/go-yaml"
	log "github.com/lukemassa/clilog"

	"filippo.io/age"
)

var topicRe = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)

type notifierType string

const (
	ntfyNotfier   notifierType = "ntfy"
	chimeNotifier notifierType = "chime"
)

type RawUnitConfiguration struct {
	Name    string                     `yaml:"name"`
	Address string                     `yaml:"address"`
	OnPress []RawNotificationMechanism `yaml:"on_press"`
}

type UnitConfiguration struct {
	Name    string
	Address string
	OnPress []NotificationMechanism
}

type RawNotificationMechanism struct {
	NtfySettings  *RawNtfySettings `yaml:"ntfy"`
	ChimeSettings *ChimeSettings   `yaml:"chime"`
}

type NotificationMechanism struct {
	NotifierType  notifierType
	NtfySettings  *NtfySettings
	ChimeSettings *ChimeSettings
}

type RawNtfySettings struct {
	EncryptedTopic string `yaml:"encryptedTopic"`
}

type NtfySettings struct {
	Topic string
}

type ChimeSettings struct {
	Address string `yaml:"address"`
}

type RawConfig struct {
	MQTTURL            string                          `yaml:"mqttURL"`
	AgeKeyFile         string                          `yaml:"ageKeyFile"`
	UnitConfigurations map[string]RawUnitConfiguration `yaml:"units"`
}

type Config struct {
	MQTTURL            string
	UnitConfigurations map[string]UnitConfiguration
}

func NewConfig(content []byte, showSecrets bool) (*Config, error) {

	var raw RawConfig

	err := yaml.Unmarshal(content, &raw)
	if err != nil {
		return nil, err
	}
	return raw.validate(showSecrets)
}

func (c *RawConfig) validate(showSecrets bool) (*Config, error) {
	if c.AgeKeyFile == "" {
		return nil, errors.New("did not set ageKeyFile")
	}

	identities, err := loadIdentities(c.AgeKeyFile)
	if err != nil {
		return nil, fmt.Errorf("loading identities: %v", err)
	}

	unitConfigurations := make(map[string]UnitConfiguration)
	for unitID, unitConfiguration := range c.UnitConfigurations {
		if len(unitConfiguration.OnPress) == 0 {
			log.Warnf("No notification mechanism set for %s", unitID)
		}
		var notificationMechanisms []NotificationMechanism

		for i, notificationConfig := range unitConfiguration.OnPress {
			mechanism, err := c.getNotificationMechanismFromRaw(notificationConfig, identities, showSecrets)
			if err != nil {
				return nil, fmt.Errorf("configuring %s notifier for %s: %v", indexToOrdinal(i), unitID, err)
			}
			notificationMechanisms = append(notificationMechanisms, mechanism)

		}
		unitConfigurations[unitID] = UnitConfiguration{
			Name:    unitConfiguration.Name,
			Address: unitConfiguration.Address,
			OnPress: notificationMechanisms,
		}
	}
	return &Config{
		MQTTURL:            c.MQTTURL,
		UnitConfigurations: unitConfigurations,
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

func (c *RawConfig) getNotificationMechanismFromRaw(notificationConfig RawNotificationMechanism, identities []age.Identity, showSecrets bool) (NotificationMechanism, error) {

	if notificationConfig.NtfySettings != nil {
		topic, err := decrypt(notificationConfig.NtfySettings.EncryptedTopic, identities)
		if err != nil {
			return NotificationMechanism{}, err
		}
		topicToShow := "<redacted>"
		if showSecrets {
			topicToShow = topic
		}
		if !topicRe.MatchString(topic) {
			return NotificationMechanism{}, fmt.Errorf("ntfy token does not match %s: %s", topicRe, topicToShow)
		}

		log.Infof("Configured ntfy notifier with topic:%s", topicToShow)
		return NotificationMechanism{
			NotifierType: ntfyNotfier,
			NtfySettings: &NtfySettings{
				Topic: topic,
			},
		}, nil
	}
	if notificationConfig.ChimeSettings != nil {
		log.Infof("Configured chime notifier with address %s", notificationConfig.ChimeSettings.Address)
		return NotificationMechanism{
			NotifierType:  chimeNotifier,
			ChimeSettings: notificationConfig.ChimeSettings,
		}, nil
	}

	return NotificationMechanism{}, errors.New("could not determine which notifier to use")

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
