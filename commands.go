package main

import (
	"fmt"
	"strings"
	"sync"
)

type Details struct {
	name              string
	arity             int
	flags             interface{}
	firstKey          int
	lastKey           int
	step              int
	aclCategories     []string
	tips              interface{} // interface{} means that field is not supported currently
	keySpecifications interface{}
	subcommands       interface{}
}

func (d *Details) ToValue() Value {
	val := Value{typ: "array", array: make([]Value, 0, 10)}
	val.array = append(val.array, MakeStringValue(d.name))
	val.array = append(val.array, MakeIntValue(d.arity))
	val.array = append(val.array, MakeNilValue())
	val.array = append(val.array, MakeIntValue(d.firstKey))
	val.array = append(val.array, MakeIntValue(d.lastKey))
	val.array = append(val.array, MakeIntValue(d.step))

	args := make([]any, len(d.aclCategories))
	for i, v := range d.aclCategories {
		args[i] = v
	}
	val.array = append(val.array, MakeArrayValue(args...))

	val.array = append(val.array, MakeNilValue())
	val.array = append(val.array, MakeNilValue())
	val.array = append(val.array, MakeNilValue())

	return val
}

type Command struct {
	details Details
	handler func([]Value) Value
}

type CommandHandler struct {
	commands map[string]Command
}

func NewCommandHandler() *CommandHandler {
	commands := map[string]Command{}

	commands["COMMAND"] = Command{
		details: Details{
			name:              "command",
			arity:             1,
			flags:             nil,
			firstKey:          0,
			lastKey:           0,
			step:              0,
			aclCategories:     []string{"@slow", "@connection"},
			tips:              nil,
			keySpecifications: nil,
			subcommands:       nil,
		},
		handler: command,
	}

	commands["PING"] = Command{
		details: Details{
			name:              "ping",
			arity:             -1,
			flags:             nil,
			firstKey:          0,
			lastKey:           0,
			step:              0,
			aclCategories:     []string{"@fast", "@connection"},
			tips:              nil,
			keySpecifications: nil,
			subcommands:       nil,
		},
		handler: ping,
	}

	commands["SET"] = Command{
		details: Details{
			name:              "set",
			arity:             3,
			flags:             nil,
			firstKey:          1,
			lastKey:           1,
			step:              1,
			aclCategories:     []string{"@write", "@string", "@fast"},
			tips:              nil,
			keySpecifications: nil,
			subcommands:       nil,
		},
		handler: set,
	}

	commands["GET"] = Command{
		details: Details{
			name:              "get",
			arity:             2,
			flags:             nil,
			firstKey:          1,
			lastKey:           1,
			step:              1,
			aclCategories:     []string{"@read", "@string", "@fast"},
			tips:              nil,
			keySpecifications: nil,
			subcommands:       nil,
		},
		handler: get,
	}

	commands["HSET"] = Command{
		details: Details{
			name:              "hset",
			arity:             4,
			flags:             nil,
			firstKey:          1,
			lastKey:           1,
			step:              1,
			aclCategories:     []string{"@write", "@hash", "@fast"},
			tips:              nil,
			keySpecifications: nil,
			subcommands:       nil,
		},
		handler: hset,
	}

	commands["HGET"] = Command{
		details: Details{
			name:              "hget",
			arity:             3,
			flags:             nil,
			firstKey:          1,
			lastKey:           1,
			step:              1,
			aclCategories:     []string{"@read", "@hash", "@fast"},
			tips:              nil,
			keySpecifications: nil,
			subcommands:       nil,
		},
		handler: hget,
	}

	commands["HGETALL"] = Command{
		details: Details{
			name:              "hgetall",
			arity:             2,
			flags:             nil,
			firstKey:          1,
			lastKey:           1,
			step:              1,
			aclCategories:     []string{"@read", "@hash", "@fast"},
			tips:              nil,
			keySpecifications: nil,
			subcommands:       nil,
		},
		handler: hgetall,
	}

	return &CommandHandler{commands: commands}
}

var SETMap = map[string]string{}
var SETMapMutex = sync.RWMutex{}

var HSETMap = map[string]map[string]string{}
var HSETMapMutex = sync.RWMutex{}

func (c *CommandHandler) Handle(command string, args []Value) (Value, error) {
	cmd, ok := c.commands[command]
	if !ok {
		return Value{}, fmt.Errorf("Invalid command: %s", command)
	}
	return cmd.handler(args), nil
}

func commandDocs() Value {
	return Value{typ: "map"}
}

func command(args []Value) Value {
	if len(args) > 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'command' command"}
	}
	if len(args) == 1 {
		if strings.ToUpper(args[0].bulk) == "DOCS" {
			return commandDocs()
		}
		return Value{typ: "error", str: "ERR invalid argument at index 1"}
	}

	cmdHandler := NewCommandHandler()
	value := Value{typ: "array", array: make([]Value, 0, len(cmdHandler.commands))}

	for _, cmd := range cmdHandler.commands {
		value.array = append(value.array, cmd.details.ToValue())
	}

	return value
}

func ping(args []Value) Value {
	if len(args) == 0 {
		return Value{typ: "string", str: "PONG"}
	}

	return Value{typ: "string", str: args[0].bulk}
}

func set(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'set' command"}
	}

	key := args[0].bulk
	val := args[1].bulk

	SETMapMutex.Lock()
	SETMap[key] = val
	SETMapMutex.Unlock()

	return Value{typ: "string", str: "OK"}
}

func get(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'get' command"}
	}

	key := args[0].bulk
	SETMapMutex.RLock()
	val, ok := SETMap[key]
	SETMapMutex.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: val}
}

func hset(args []Value) Value {
	if len(args) != 3 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hset' command"}
	}

	hash := args[0].bulk
	key := args[1].bulk
	val := args[2].bulk

	HSETMapMutex.Lock()
	if _, ok := HSETMap[hash]; !ok {
		HSETMap[hash] = make(map[string]string)
	}
	HSETMap[hash][key] = val
	HSETMapMutex.Unlock()

	return Value{typ: "string", str: "OK"}
}

func hget(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hget' command"}
	}

	hash := args[0].bulk
	key := args[1].bulk

	HSETMapMutex.RLock()
	defer HSETMapMutex.RUnlock()

	if _, ok := HSETMap[hash]; !ok {
		return Value{typ: "null"}
	}
	val, ok := HSETMap[hash][key]

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: val}
}

func hgetall(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments for 'hgetall' command"}
	}

	hash := args[0].bulk
	HSETMapMutex.RLock()
	defer HSETMapMutex.RUnlock()

	if _, ok := HSETMap[hash]; !ok {
		return Value{typ: "null"}
	}

	result := make([]Value, 0, len(HSETMap[hash])*2)
	for key, val := range HSETMap[hash] {
		result = append(result, Value{typ: "bulk", bulk: key})
		result = append(result, Value{typ: "bulk", bulk: val})
	}

	return Value{typ: "array", array: result}
}
