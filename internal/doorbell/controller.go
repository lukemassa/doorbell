package doorbell

import (
	"log"
	"time"
)

const baseHealthURL = "https://hc-ping.com/4003a09f-f033-4f38-82ff-a6a0f010fa50"
const updateFreq = 10 * time.Minute
const maxTemp = 55 // degrees celsius

type Controller struct {
	mqttURL string
	units   []Unit
}

func (c Controller) LookupUnit(lookup string) (Unit, bool) {
	for _, unit := range c.units {
		if unit.ID == lookup || unit.Address == lookup {
			log.Printf("Found unit %s for lookup %s", unit.Name, lookup)
			return unit, true
		}
	}
	return Unit{}, false
}

func (c *Controller) Run() error {

	ringChan, err := c.subscribe()
	if err != nil {
		return err
	}

	healthCheckTimer := time.NewTicker(updateFreq)
	defer healthCheckTimer.Stop()

	c.updateSystemHealth()
	for {
		select {
		case <-healthCheckTimer.C:
			c.updateSystemHealth()
		case bellPress := <-ringChan:
			c.Ring(bellPress)
		}
	}
}
