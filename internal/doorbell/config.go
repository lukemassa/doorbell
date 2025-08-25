package doorbell

import (
	"errors"
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	MQTTURL             string `yaml:"mqttURL"`
	NtfyTopicSuffixFile string `yaml:"ntfyTopicSuffixFile"`
}

func NewConfig(content []byte) (*Config, error) {

	var ret Config

	err := yaml.Unmarshal(content, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Config) Validate() error {
	if c.NtfyTopicSuffixFile == "" {
		return errors.New("did not set ntfyTopicSuffixFile")
	}
	_, err := os.Stat(c.NtfyTopicSuffixFile)
	if err != nil {
		return err
	}
	return nil
}
