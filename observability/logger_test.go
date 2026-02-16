package observability

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"sync"
	"testing"
)

func newTestLogger(level LogLevel) (*Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	l := &Logger{
		level:  level,
		logger: log.New(buf, "", 0), // no timestamps for predictable output
	}
	return l, buf
}

func TestNewLogger_DefaultLevelFallback(t *testing.T) {
	l := NewLogger(LogLevel(999)) // invalid level
	if l.level != LogInfo {
		t.Errorf("expected default level LogInfo, got %v", l.level)
	}
}

func TestLogLevelFiltering(t *testing.T) {
	l, buf := newTestLogger(LogWarn)

	l.Info("info message")
	if buf.Len() != 0 {
		t.Errorf("expected Info not to log at Warn level")
	}

	l.Warn("warn message")
	if !strings.Contains(buf.String(), "WARN warn message") {
		t.Errorf("expected Warn to be logged")
	}
}

func TestStructuredLogging(t *testing.T) {
	l, buf := newTestLogger(LogDebug)

	l.Info("user_login", "user", "alice", "id", 123)

	out := buf.String()

	if !strings.Contains(out, "INFO user_login") {
		t.Errorf("missing level or message")
	}
	if !strings.Contains(out, "user=alice") {
		t.Errorf("missing structured key=value pair")
	}
	if !strings.Contains(out, "id=123") {
		t.Errorf("missing numeric key=value pair")
	}
}

func TestStringify(t *testing.T) {
	if stringify("abc") != "abc" {
		t.Errorf("string stringify failed")
	}

	err := errors.New("boom")
	if stringify(err) != "boom" {
		t.Errorf("error stringify failed")
	}

	if stringify(123) != "123" {
		t.Errorf("int stringify failed")
	}
}

func TestOddKeyValuePairsIgnored(t *testing.T) {
	l, buf := newTestLogger(LogInfo)

	// odd number of kv arguments — last should be ignored
	l.Info("test", "key1", "value1", "lonely_key")

	out := buf.String()

	if !strings.Contains(out, "key1=value1") {
		t.Errorf("expected valid pair to appear")
	}
	if strings.Contains(out, "lonely_key=") {
		t.Errorf("unexpected dangling key logged")
	}
}

func TestConcurrentLogging(t *testing.T) {
	l, buf := newTestLogger(LogDebug)

	wg := sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			l.Info("msg", "i", i)
		}(i)
	}

	wg.Wait()

	out := buf.String()

	// Just ensure we logged many lines without panic
	if len(strings.Split(strings.TrimSpace(out), "\n")) < 50 {
		t.Errorf("expected multiple log lines from concurrent logging")
	}
}
