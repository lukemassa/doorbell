package health

import (
	"bytes"
	"fmt"
	"io"
	"time"

	log "github.com/lukemassa/clilog"

	"github.com/hashicorp/go-retryablehttp"

	"encoding/json"
)

const baseHealthURL = "https://hc-ping.com/4003a09f-f033-4f38-82ff-a6a0f010fa50"

const updateFreq = 10 * time.Minute

type SystemStatus struct {
	temp float64
	err  error
}

type SystemReport struct {
	Temp    float64 `json:"temp"`
	Message string  `json:"message"`
	OK      bool    `json:"ok"`
}

func NewSystemStatus() *SystemStatus {
	return &SystemStatus{}
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
	if s.temp < minTemp {
		ret.Message = fmt.Sprintf("Temp below threshold %dC", maxTemp)
		ret.OK = false
		return ret
	}
	ret.OK = true
	ret.Message = "All systems normal"
	return ret

}

func (s *SystemStatus) Run() {
	go func() {
		healthCheckTimer := time.NewTicker(updateFreq)
		s.runHealthcheck()

		for range healthCheckTimer.C {
			s.runHealthcheck()
		}
	}()
}

func (s *SystemStatus) updateSystemHealth() {
	temp, err := readCPUTemp()
	if err != nil {
		s.err = err
		return
	}
	s.temp = temp
}

func (s *SystemStatus) runHealthcheck() {

	s.updateSystemHealth()

	url := baseHealthURL

	report := s.Report()
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
