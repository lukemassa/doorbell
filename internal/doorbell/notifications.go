package doorbell

import (
	"fmt"
	"net/http"
	"strings"
)

type Notifier interface {
	Notify() error
}

type NtfyNotifier struct {
	topic   string
	message string
}

func (r *NtfyNotifier) Notify() error {
	_, err := http.Post(r.url(), "text/plain", strings.NewReader(r.message))
	if err != nil {
		return fmt.Errorf("notifying ntfy: %v", err)
	}
	return nil
}

func (r NtfyNotifier) url() string {
	return fmt.Sprintf("https://ntfy.sh/%s", r.topic)
}
