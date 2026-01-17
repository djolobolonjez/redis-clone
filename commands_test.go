package main

import (
	"strings"
	"testing"
)

func runCommand(input Value) (Value, error) {
	cmdHandler := NewCommandHandler()
	cmd := strings.ToUpper(input.array[0].bulk)
	return cmdHandler.Handle(cmd, input.array[1:])
}

func TestCommandPing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "just ping",
			input:    "*1\r\n$4\r\nPING\r\n",
			expected: "+PONG\r\n",
		},
		{
			name:     "ping with argument",
			input:    "*2\r\n$4\r\nPING\r\n$11\r\nhello world\r\n",
			expected: "+hello world\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := RespReaderFromString(tt.input)
			output, err := resp.Read()
			if err != nil {
				t.Fatal(err)
			}

			if output.typ != "array" {
				t.Fatal("Invalid input, expected array")
			}

			if len(output.array) == 0 {
				t.Fatal("Invalid input, array length must be greater than zero")
			}

			val, err := runCommand(output)
			if err != nil {
				t.Fatal(err)
			}

			result := string(val.Serialize())
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
