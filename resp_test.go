package main

import (
	"testing"
)

func TestRespReadInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{name: "positive integer", input: ":100\r\n", expected: 100},
		{name: "positive integer with + sign", input: ":+100\r\n", expected: 100},
		{name: "negative integer", input: ":-100\r\n", expected: -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := RespReaderFromString(tt.input)
			output, err := resp.Read()
			if err != nil {
				t.Fatal(err)
			}

			if output.num != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, output.num)
			}
		})
	}
}

func TestRespReadBulk(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "hello string", input: "$5\r\nhello\r\n", expected: "hello"},
		{name: "hello world string", input: "$11\r\nhello world\r\n", expected: "hello world"},
		{name: "empty string", input: "$0\r\n\r\n", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := RespReaderFromString(tt.input)
			output, err := resp.Read()
			if err != nil {
				t.Fatal(err)
			}

			if output.bulk != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, output.bulk)
			}
		})
	}
}

func TestRespReadArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Value
	}{
		{name: "empty array", input: "*0\r\n", expected: Value{}},
		{
			name:  "array of two strings",
			input: "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			expected: Value{
				typ: "array",
				array: []Value{
					{
						typ:  "bulk",
						bulk: "hello",
					},
					{
						typ:  "bulk",
						bulk: "world",
					},
				},
			},
		},
		{
			name:  "array of two integers",
			input: "*2\r\n:1\r\n:-2\r\n",
			expected: Value{
				typ: "array",
				array: []Value{
					{
						typ: "integer",
						num: 1,
					},
					{
						typ: "integer",
						num: -2,
					},
				},
			},
		},
		{
			name:  "array of string and integer",
			input: "*2\r\n$5\r\nhello\r\n:-2\r\n",
			expected: Value{
				typ: "array",
				array: []Value{
					{
						typ:  "bulk",
						bulk: "hello",
					},
					{
						typ: "integer",
						num: -2,
					},
				},
			},
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
				t.Fatalf("expected array type, got %s", output.typ)
			}

			for i := 0; i < len(output.array); i++ {
				expType := tt.expected.array[i].typ
				outType := output.array[i].typ
				if expType != outType {
					t.Errorf("expected %s element at index %d, got %s", expType, i, outType)
					continue
				}
				switch expType {
				case "integer":
					expNum := tt.expected.array[i].num
					outNum := output.array[i].num
					if outNum != expNum {
						t.Errorf("expected %d at index %d, got %d", expNum, i, outNum)
					}
				case "bulk":
					expBulk := tt.expected.array[i].bulk
					outBulk := output.array[i].bulk
					if outBulk != expBulk {
						t.Errorf("expected %s at index %d, got %s", expBulk, i, outBulk)
					}
				}
			}
		})

	}
}

func TestRespWriteNull(t *testing.T) {
	val := Value{typ: "null"}
	result := string(val.Serialize())
	expected := "$-1\r\n"

	if result != expected {
		t.Errorf("Invalid null value, expected %s, got %s", expected, result)
	}
}

func TestRespWriteInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    Value
		expected string
	}{
		{name: "positive integer", input: Value{typ: "integer", num: 100}, expected: ":100\r\n"},
		{name: "negative integer", input: Value{typ: "integer", num: -100}, expected: ":-100\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(tt.input.Serialize())

			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRespWriteBulk(t *testing.T) {
	tests := []struct {
		name     string
		input    Value
		expected string
	}{
		{name: "hello", input: Value{typ: "bulk", bulk: "hello"}, expected: "$5\r\nhello\r\n"},
		{name: "hello world", input: Value{typ: "bulk", bulk: "hello world"}, expected: "$11\r\nhello world\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(tt.input.Serialize())

			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRespWriteString(t *testing.T) {
	tests := []struct {
		name     string
		input    Value
		expected string
	}{
		{name: "hello", input: Value{typ: "string", str: "OK"}, expected: "+OK\r\n"},
		{name: "hello world", input: Value{typ: "string", str: "hello world"}, expected: "+hello world\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(tt.input.Serialize())

			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRespWriteArray(t *testing.T) {
	tests := []struct {
		name     string
		input    Value
		expected string
	}{
		{
			name: "array of two strings",
			input: Value{
				typ: "array",
				array: []Value{
					{
						typ:  "bulk",
						bulk: "hello",
					},
					{
						typ:  "bulk",
						bulk: "world",
					},
				},
			},
			expected: "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
		},
		{
			name: "array of two integers",
			input: Value{
				typ: "array",
				array: []Value{
					{
						typ: "integer",
						num: 100,
					},
					{
						typ: "integer",
						num: -100,
					},
				},
			},
			expected: "*2\r\n:100\r\n:-100\r\n",
		},
		{
			name: "array of integer and a string",
			input: Value{
				typ: "array",
				array: []Value{
					{
						typ: "integer",
						num: -100,
					},
					{
						typ:  "bulk",
						bulk: "hello world",
					},
				},
			},
			expected: "*2\r\n:-100\r\n$11\r\nhello world\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(tt.input.Serialize())

			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRespWriteError(t *testing.T) {
	val := Value{typ: "error", str: "ERR wrong number of arguments"}
	result := string(val.Serialize())
	expected := "-ERR wrong number of arguments\r\n"

	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}
