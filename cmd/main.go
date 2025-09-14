package main

import (
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	log "github.com/lukemassa/clilog"
	"github.com/lukemassa/doorbell/internal/doorbell"
)

const (
	defaultConfigPath = "conf/config.yaml"
)

type ValidateCommand struct {
	ShowSecrets bool `long:"show-secrets" description:"Show Secrets"`
}

type RingCommand struct {
	UnitID string `long:"unit-id" short:"u" description:"Unit ID" required:"true"`
}

type RunCommand struct {
}

func (v *ValidateCommand) Execute(args []string) error {
	mustGetController(v.ShowSecrets)
	log.Info("Valid config")
	return nil
}

func (r *RingCommand) Execute(args []string) error {

	controller := mustGetController(false)
	controller.SetOnMock(func(message string) {
		log.Infof("Mock received message: %s", message)
	})
	controller.Ring(doorbell.BellPress{
		UnitID: r.UnitID,
		Action: "single",
	})
	// TODO: Better way to wait
	time.Sleep(1 * time.Second)
	return nil
}

func (r *RunCommand) Execute(args []string) error {

	controller := mustGetController(false)

	err := controller.Run()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func mustAddCommand(parser *flags.Parser, name, shortDesc, longDesc string, command any) *flags.Command {
	cmd, err := parser.AddCommand(name, shortDesc, longDesc, command)
	if err != nil {
		log.Fatalf("failed to add %q command: %v", name, err)
	}
	return cmd
}

func getConfigContent() ([]byte, error) {
	configPath := os.Getenv("DOORBELL_CONFIG")
	if configPath == "" {
		configPath = defaultConfigPath
	}
	return os.ReadFile(configPath)
}

func mustGetController(showSecrets bool) *doorbell.Controller {
	content, err := getConfigContent()
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := doorbell.NewConfig(content, showSecrets)
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
		log.MustSetFormat(`{{ .Level }} {{ .Message }}`)
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
