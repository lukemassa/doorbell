package doorbell

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/lukemassa/clilog"
)

const baseHealthURL = "https://hc-ping.com/4003a09f-f033-4f38-82ff-a6a0f010fa50"
const updateFreq = 10 * time.Minute

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

	shutDown := make(chan os.Signal, 1)
	signal.Notify(shutDown, syscall.SIGINT, syscall.SIGTERM)

	healthCheckTimer := time.NewTicker(updateFreq)
	defer healthCheckTimer.Stop()

	c.updateSystemHealth()
	for {
		select {
		case <-healthCheckTimer.C:
			c.updateSystemHealth()
		case bellPress := <-client.bellPressChan:
			c.Ring(bellPress)
		case <-shutDown:
			log.Info("Shutting down...")
			client.Shutdown()
			return nil
		}
	}
}
