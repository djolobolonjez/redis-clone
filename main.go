package main

import (
	"io"
	"log"
	"net"
	"strings"
)

func readLoop(conn net.Conn, aof *Aof) {
	defer conn.Close()

	cmdHandler := NewCommandHandler()

	for {
		reader := NewRespReader(conn)
		request, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Println("error reading from client:", err)
			return
		}

		if request.typ != "array" {
			log.Println("Invalid request, expected array")
			continue
		}

		if len(request.array) == 0 {
			log.Println("Invalid request, array length must be greater than zero")
			continue
		}

		command := strings.ToUpper(request.array[0].bulk)
		args := request.array[1:]

		writer := NewRespWriter(conn)

		result, err := cmdHandler.Handle(command, args)
		if err != nil {
			log.Println(err)
			writer.Write(Value{typ: "string", str: ""})
			continue
		}

		if command == "SET" || command == "HSET" {
			aof.Write(request)
		}

		writer.Write(result)
	}
}

func main() {

	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		log.Fatal(err)
	}

	aof, err := NewAof("resplog.aof")
	if err != nil {
		log.Fatal(err)
	}

	defer aof.Close()

	aof.Read(func(value Value) {
		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]
		cmdHandler := NewCommandHandler()

		_, err := cmdHandler.Handle(command, args)
		if err != nil {
			log.Fatal(err)
		}
	})

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("error in accept: ", err)
			continue
		}

		go readLoop(conn, aof)
	}
}
