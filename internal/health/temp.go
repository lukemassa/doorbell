package health

import (
	"os"
	"strconv"
	"strings"
)

const (
	maxTemp = 55 // degrees celsius
	minTemp = 15
)

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
