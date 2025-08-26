package doorbell

import (
	"errors"
	"os"

	"github.com/goccy/go-yaml"

	"filippo.io/age"
)

type UnitConfiguration struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
}

type Config struct {
	MQTTURL             string                       `yaml:"mqttURL"`
	AgeKeyFile          string                       `yaml:"ageKeyFile"`
	UnitConfigurations  map[string]UnitConfiguration `yaml:"units"`
	EncryptedNtfySuffix string                       `yaml:"encryptedNtfySuffix"`
	ageIdentities       []age.Identity
}

func NewConfig(content []byte) (*Config, error) {

	var ret Config

	err := yaml.Unmarshal(content, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *Config) Load() error {
	if c.AgeKeyFile == "" {
		return errors.New("did not set ageKeyFile")
	}
	identities, err := loadIdentities(c.AgeKeyFile)
	if err != nil {
		return err
	}
	c.ageIdentities = identities
	return nil
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
