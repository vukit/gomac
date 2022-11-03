package logger_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/vukit/gomac/internal/gophermart/logger"
)

var message struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

func TestInfo(t *testing.T) {
	var err error

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	tLogger := logger.NewLogger(w)

	tLogger.Info("info message")

	buffer := make([]byte, 1024)
	n, err := r.Read(buffer)
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(buffer[:n], &message)
	if err != nil {
		t.Fatal(err)
	}

	if message.Level != "info" {
		t.Errorf("wrong level, want 'info', got '%s'", message.Level)
	}

	if message.Message != "info message" {
		t.Errorf("wrong message, want 'info message', got '%s'", message.Message)
	}
}

func TestWarning(t *testing.T) {
	var err error

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	tLogger := logger.NewLogger(w)

	tLogger.Warning("warning message")

	buffer := make([]byte, 1024)
	n, err := r.Read(buffer)
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(buffer[:n], &message)
	if err != nil {
		t.Fatal(err)
	}

	if message.Level != "warn" {
		t.Errorf("wrong level, want 'warning', got '%s'", message.Level)
	}

	if message.Message != "warning message" {
		t.Errorf("wrong msg, want 'warning message', got '%s'", message.Message)
	}
}
