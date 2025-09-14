package doorbell

import (
	"fmt"
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

func (c *Config) Controller() (*Controller, error) {

	var units []Unit
	for unitID, unitConfiguration := range c.UnitConfigurations {
		var notifiers []Notifier

		for _, notificationConfig := range unitConfiguration.OnPress {
			var notifier Notifier
			switch notificationConfig.NotifierType {
			case ntfyNotfier:
				notifier = NtfyNotifier{
					topic:   notificationConfig.NtfySettings.Topic,
					message: fmt.Sprintf("Ring for %s", unitConfiguration.Name),
				}
			case chimeNotifier:
				notifier = ChimeNotifier{
					mqttURL: c.MQTTURL,
					address: notificationConfig.ChimeSettings.Address,
				}
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
