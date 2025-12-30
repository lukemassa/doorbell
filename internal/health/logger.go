package health

import (
	"fmt"

	log "github.com/lukemassa/clilog"
)

// Simple logger to use in the retryable client
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
