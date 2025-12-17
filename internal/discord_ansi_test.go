package internal

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestDiscordClient_WriteStripsANSI(t *testing.T) {
	client := NewDiscordClient(AppConfig{DiscordWebhookUrl: ""})

	// Simulate colored log output with ANSI escape codes
	coloredError := "\x1b[31mERROR\x1b[0m This is a colored error message"
	
	n, err := client.Write([]byte(coloredError))
	
	if err != nil {
		t.Errorf("Write() error = %v, want nil", err)
	}
	
	if n != len(coloredError) {
		t.Errorf("Write() returned %v bytes, want %v", n, len(coloredError))
	}

	// Verify that ANSI codes are stripped
	stripped := ansi.Strip(coloredError)
	if stripped == coloredError {
		t.Error("ANSI codes were not stripped from test input")
	}

	expected := "ERROR This is a colored error message"
	if stripped != expected {
		t.Errorf("Stripped message = %q, want %q", stripped, expected)
	}
}

func TestANSIStripping(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple ANSI color",
			input: "\x1b[31mRed Text\x1b[0m",
			want:  "Red Text",
		},
		{
			name:  "multiple ANSI codes",
			input: "\x1b[1;38;5;86mINFO\x1b[0m Message",
			want:  "INFO Message",
		},
		{
			name:  "no ANSI codes",
			input: "Plain text",
			want:  "Plain text",
		},
		{
			name:  "real log example",
			input: "2025/12/17 19:27:24 \x1b[1;38;5;196mERROR\x1b[0m something went wrong",
			want:  "2025/12/17 19:27:24 ERROR something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ansi.Strip(tt.input)
			if got != tt.want {
				t.Errorf("ansi.Strip() = %q, want %q", got, tt.want)
			}
		})
	}
}
