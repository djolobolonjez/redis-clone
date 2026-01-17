package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeCommand(parts ...string) Value {
	arr := make([]Value, 0, len(parts))
	for _, p := range parts {
		arr = append(arr, Value{typ: "bulk", bulk: p})
	}
	return Value{typ: "array", array: arr}
}

func TestAofPersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.aof")

	aof, err := NewAof(path)
	if err != nil {
		t.Fatalf("failed to create aof: %v", err)
	}

	setCmd := makeCommand("SET", "key1", "value1")
	hsetCmd := makeCommand("HSET", "myhash", "field1", "hvalue1")

	if err := aof.Write(setCmd); err != nil {
		aof.Close()
		t.Fatalf("failed to write set command to aof: %v", err)
	}
	if err := aof.Write(hsetCmd); err != nil {
		aof.Close()
		t.Fatalf("failed to write hset command to aof: %v", err)
	}

	if err := aof.Close(); err != nil {
		t.Fatalf("failed to close aof: %v", err)
	}

	aof2, err := NewAof(path)
	if err != nil {
		t.Fatalf("failed to open aof: %v", err)
	}

	defer aof2.Close()
	defer os.Remove(path)

	cmdHandler := NewCommandHandler()

	err = aof2.Read(func(value Value) {
		if value.typ != "array" || len(value.array) == 0 {
			t.Error("Invalid data")
			return
		}
		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		_, err := cmdHandler.Handle(command, args)
		if err != nil {
			t.Error(err)
		}
	})
	if err != nil {
		t.Fatalf("failed to read aof: %v", err)
	}

	if v, ok := SETMap["key1"]; !ok || v != "value1" {
		t.Fatalf("expected SET key1=value1, got %v (exists=%v)", v, ok)
	}

	if _, ok := HSETMap["myhash"]; !ok {
		t.Fatalf("expected HSET map for 'myhash' to exist")
	}
	if hv, ok := HSETMap["myhash"]["field1"]; !ok || hv != "hvalue1" {
		t.Fatalf("expected HSET myhash.field1=hvalue1, got %v (exists=%v)", hv, ok)
	}
}
