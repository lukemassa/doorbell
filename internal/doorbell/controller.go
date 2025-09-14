package doorbell

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/lukemassa/clilog"
)

const baseHealthURL = "https://hc-ping.com/4003a09f-f033-4f38-82ff-a6a0f010fa50"

type Controller struct {
	mqttURL string
	units   []Unit
}

func (c Controller) LookupUnit(lookup string) (Unit, bool) {
	for _, unit := range c.units {
		if unit.ID == lookup || unit.Address == lookup {
			log.Infof("Found unit %s for lookup %s", unit.Name, lookup)
			return unit, true
		}
	}
	return Unit{}, false
}

func (c *Controller) Run() error {

	client, err := c.subscribe()
	if err != nil {
		return err
	}

	systemStatus := newSystemStatus()
	systemStatus.Run()

	shutDown := make(chan os.Signal, 1)
	signal.Notify(shutDown, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case bellPress := <-client.bellPressChan:
			// TODO: If bell press has some info for our system status, let it know!
			c.Ring(bellPress)
		case <-shutDown:
			log.Info("Shutting down...")
			client.Shutdown()
			return nil
		}
	}
}

func (c *Controller) Validate(showSecrets bool) error {
	for _, unit := range c.units {
		log.Infof("Validating unit: %s", unit.Name)
		for _, notifier := range unit.Notifiers {
			err := notifier.Validate(showSecrets)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
