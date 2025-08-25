package main

import (
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/lukemassa/doorbell/internal/doorbell"
)

const (
	defaultConfigPath = "conf/test.yaml"
)

type ValidateCommand struct {
}

type RingCommand struct {
}

type RunCommand struct {
}

func (f *ValidateCommand) Execute(args []string) error {
	mustGetConfig()
	log.Print("Valid config")
	return nil
}

func (l *RingCommand) Execute(args []string) error {

	cfg := mustGetConfig()

	controller, err := doorbell.NewController(cfg)
	if err != nil {
		log.Fatal(err)
	}
	controller.Ring(doorbell.BellPress{
		Unit:   "first_floor",
		Action: "single",
	})
	return nil
}

func (r *RunCommand) Execute(args []string) error {

	cfg := mustGetConfig()

	controller, err := doorbell.NewController(cfg)
	if err != nil {
		log.Fatal(err)
	}
	err = controller.Run()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func main() {

	parser := flags.NewParser(nil, flags.Default)
	parser.AddCommand("validate", "Validates config", "Makes sure that the config file is valid", &ValidateCommand{})
	parser.AddCommand("ring", "Rings a unit", "Acts as if the given unit has been rung", &RingCommand{})
	parser.AddCommand("run", "Runs the controller", "Runs the controller, responding to the doorbell rings", &RunCommand{})
	_, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}
}

func getConfigContent() ([]byte, error) {
	configPath := os.Getenv("DOORBELL_CONFIG")
	if configPath == "" {
		configPath = defaultConfigPath
	}
	return os.ReadFile(configPath)
}

func mustGetConfig() *doorbell.Config {
	content, err := getConfigContent()
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := doorbell.NewConfig(content)
	if err != nil {
		log.Fatal(err)
	}
	err = cfg.Validate()
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}
