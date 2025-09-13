package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/lukemassa/clilog"
	log "github.com/lukemassa/clilog"
	"github.com/lukemassa/doorbell/internal/doorbell"
)

const (
	defaultConfigPath = "conf/config.yaml"
)

type ValidateCommand struct {
}

type RingCommand struct {
}

type RunCommand struct {
}

func (f *ValidateCommand) Execute(args []string) error {
	mustGetController()
	log.Info("Valid config")
	return nil
}

func (l *RingCommand) Execute(args []string) error {

	controller := mustGetController()
	if len(args) != 1 {
		return fmt.Errorf("unexpected number of args: %d", len(args))
	}
	unitID := args[0]
	controller.Ring(doorbell.BellPress{
		UnitID: unitID,
		Action: "single",
	})
	return nil
}

func (r *RunCommand) Execute(args []string) error {

	controller := mustGetController()

	err := controller.Run()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func mustAddCommand(parser *flags.Parser, name, shortDesc, longDesc string, command interface{}) {
	if _, err := parser.AddCommand(name, shortDesc, longDesc, command); err != nil {
		log.Fatalf("failed to add %q command: %v", name, err)
	}
}

func getConfigContent() ([]byte, error) {
	configPath := os.Getenv("DOORBELL_CONFIG")
	if configPath == "" {
		configPath = defaultConfigPath
	}
	return os.ReadFile(configPath)
}

func mustGetController() *doorbell.Controller {
	content, err := getConfigContent()
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := doorbell.NewConfig(content)
	if err != nil {
		log.Fatal(err)
	}
	controller, err := cfg.Controller()
	if err != nil {
		log.Fatal(err)
	}
	return controller
}

func main() {
	if _, ok := os.LookupEnv("JOURNAL_STREAM"); ok {
		// logs go straight to journald, so we don't need the time
		clilog.SetFormat(`{{ .Level }} {{ .Message }}`)
	}
	parser := flags.NewParser(nil, flags.Default)
	mustAddCommand(parser, "validate", "Validates config", "Makes sure that the config file is valid", &ValidateCommand{})
	mustAddCommand(parser, "ring", "Rings a unit", "Acts as if the given unit has been rung", &RingCommand{})
	mustAddCommand(parser, "run", "Runs the controller", "Runs the controller, responding to the doorbell rings", &RunCommand{})
	_, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}
}
