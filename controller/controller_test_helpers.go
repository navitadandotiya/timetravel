package controller

import "github.com/rainbowmga/timetravel/observability"

// Mock logger implements LoggerInterface
type mockLogger struct {
	Errors []string
	Infos  []string
}

func (m *mockLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (m *mockLogger) Info(msg string, keysAndValues ...interface{})  { m.Infos = append(m.Infos, msg) }
func (m *mockLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (m *mockLogger) Error(msg string, keysAndValues ...interface{}) { m.Errors = append(m.Errors, msg) }

// setLoggerForTest temporarily swaps DefaultLogger for testing
func setLoggerForTest(l observability.LoggerInterface) func() {
	orig := observability.DefaultLogger
	observability.DefaultLogger = l
	return func() { observability.DefaultLogger = orig }
}