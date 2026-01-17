package main

import (
	"bufio"
	"fmt"
	"strings"
)

func RespReaderFromString(input string) *RespReader {
	return NewRespReader(bufio.NewReader(strings.NewReader(input)))
}

func MakeIntValue(num int) Value {
	return Value{typ: "integer", num: num}
}

func MakeStringValue(str string) Value {
	return Value{typ: "string", str: str}
}

func MakeBulkValue(bulk string) Value {
	return Value{typ: "bulk", bulk: bulk}
}

func MakeArrayValue(values ...any) Value {
	val := Value{typ: "array", array: make([]Value, 0, len(values))}
	for _, v := range values {
		switch x := v.(type) {
		case int:
			val.array = append(val.array, MakeIntValue(x))
		case string:
			val.array = append(val.array, MakeStringValue(x))
		default:
			panic(fmt.Sprintf("unsupported type: %T", x))
		}
	}

	return val
}

func MakeNilValue() Value {
	return Value{typ: "null"}
}
