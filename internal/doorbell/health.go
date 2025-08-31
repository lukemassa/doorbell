package doorbell

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

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
		log.Printf("Failed to marshal json to %v", err)
	}
	log.Printf("Writing to status report: %+v", report)

	// Wrap in an io.Reader
	var r io.Reader = bytes.NewReader(b)

	resp, err := retryablehttp.Post(url, "JSON", r)
	if err != nil {
		log.Printf("Failed to post to %s: %v", url, err)
		return
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("failed to close: %v", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read body from %s: %v", url, err)
		return
	}

	fmt.Printf("Posted to %s: %s\n", url, string(body))
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
