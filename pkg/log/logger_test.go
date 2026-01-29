package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(DEBUG, &buf)
	if l == nil {
		t.Error("NewLogger returned nil")
	}
	if !l.IsLevelEnabled(DEBUG) {
		t.Error("DEBUG level should be enabled")
	}
}

func TestLogger_Log(t *testing.T) {
	tests := []struct {
		name       string
		level      Level
		checkLevel Level
		message    string
		want       string
		shouldLog  bool
	}{
		{
			name:       "debug log enabled",
			level:      DEBUG,
			checkLevel: DEBUG,
			message:    "debug message",
			want:       "[DEBUG]",
			shouldLog:  true,
		},
		{
			name:       "debug log disabled",
			level:      WARN,
			checkLevel: DEBUG,
			message:    "debug message",
			want:       "",
			shouldLog:  false,
		},
		{
			name:       "warn log enabled",
			level:      WARN,
			checkLevel: WARN,
			message:    "warn message",
			want:       "[WARN]",
			shouldLog:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := NewLogger(tt.level, &buf)
			l.Log(tt.checkLevel, tt.message)

			got := buf.String()
			if tt.shouldLog {
				if !strings.Contains(got, tt.want) {
					t.Errorf("Log() got = %q, want to contain %q", got, tt.want)
				}
				if !strings.Contains(got, tt.message) {
					t.Errorf("Log() got = %q, want to contain %q", got, tt.message)
				}
			} else if got != "" {
				t.Errorf("Log() got = %q, want empty", got)
			}
		})
	}
}

func TestLevel_Prefix(t *testing.T) {
	tests := []struct {
		name    string
		level   Level
		want    string
		wantErr bool
	}{
		{
			name:    "DEBUG",
			level:   DEBUG,
			want:    "[DEBUG] ",
			wantErr: false,
		},
		{
			name:    "WARN",
			level:   WARN,
			want:    "[WARN] ",
			wantErr: false,
		},
		{
			name:    "ERROR",
			level:   ERROR,
			want:    "[ERROR] ",
			wantErr: false,
		},
		{
			name:    "FATAL",
			level:   FATAL,
			want:    "[FATAL] ",
			wantErr: false,
		},
		{
			name:    "INVALID",
			level:   Level(99),
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.level.Prefix()
			if (err != nil) != tt.wantErr {
				t.Errorf("Level.Prefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Level.Prefix() = %v, want %v", got, tt.want)
			}
		})
	}
}
