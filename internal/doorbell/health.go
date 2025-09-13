package doorbell

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	log "github.com/lukemassa/clilog"

	"github.com/hashicorp/go-retryablehttp"

	"encoding/json"
)

const maxTemp = 55 // degrees celsius

type SystemStatus struct {
	temp float64
	err  error
}

type SystemReport struct {
	Temp    float64 `json:"temp"`
	Message string  `json:"message"`
	OK      bool    `json:"ok"`
}

func (s SystemStatus) Report() SystemReport {
	ret := SystemReport{
		Temp: s.temp,
	}
	if s.err != nil {
		ret.Message = fmt.Sprintf("ERROR: %v", s.err)
		ret.OK = false
		return ret
	}
	if s.temp > maxTemp {
		ret.Message = fmt.Sprintf("Temp above threshold %dC", maxTemp)
		ret.OK = false
		return ret
	}
	ret.OK = true
	ret.Message = "All systems normal"
	return ret

}

func (c *Controller) systemHealth() SystemStatus {
	ret := SystemStatus{}
	temp, err := readCPUTemp()
	if err != nil {
		return SystemStatus{
			err: err,
		}
	}
	ret.temp = temp
	return ret
}

func (c *Controller) updateSystemHealth() {
	health := c.systemHealth()
	updateHealthcheck(health)
}

func updateHealthcheck(status SystemStatus) {
	url := baseHealthURL

	report := status.Report()
	if !report.OK {
		url += "/fail"
	}

	// Encode to JSON
	b, err := json.Marshal(report)
	if err != nil {
		log.Errorf("Failed to marshal json to %v", err)
		return
	}
	log.Infof("Writing to status report: %+v", report)

	// Wrap in an io.Reader
	var r io.Reader = bytes.NewReader(b)
	retryableClient := retryablehttp.NewClient()
	retryableClient.Logger = cliLogLogger{}
	resp, err := retryableClient.Post(url, "JSON", r)
	if err != nil {
		log.Errorf("Failed to post to %s: %v", url, err)
		return
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Errorf("failed to close: %v", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read body from %s: %v", url, err)
		return
	}

	log.Infof("Posted to %s: %s\n", url, string(body))
}

func readCPUTemp() (float64, error) {
	data, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(data))
	milli, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	// Convert millidegrees Celsius → degrees
	return float64(milli) / 1000.0, nil
}

type cliLogLogger struct {
}

func (c cliLogLogger) fmt(msg string, keysAndValues ...any) string {
	return fmt.Sprintf("RETRYABLE %s %v", msg, keysAndValues)
}
func (c cliLogLogger) Error(msg string, keysAndValues ...any) {
	log.Error(c.fmt(msg, keysAndValues...))
}
func (c cliLogLogger) Info(msg string, keysAndValues ...any) {
	log.Info(c.fmt(msg, keysAndValues...))
}
func (c cliLogLogger) Debug(msg string, keysAndValues ...any) {
	log.Debug(c.fmt(msg, keysAndValues...))
}
func (c cliLogLogger) Warn(msg string, keysAndValues ...any) {
	log.Warn(c.fmt(msg, keysAndValues...))
}
