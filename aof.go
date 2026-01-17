package main

import (
	"io"
	"os"
	"sync"
	"time"
)

type Aof struct {
	file  *os.File
	mutex sync.Mutex
}

func NewAof(path string) (*Aof, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	aof := &Aof{
		file: f,
	}

	go func() {
		for {
			aof.mutex.Lock()
			aof.file.Sync()
			aof.mutex.Unlock()
			time.Sleep(time.Second)
		}
	}()

	return aof, nil
}

func (aof *Aof) Close() error {
	aof.mutex.Lock()
	defer aof.mutex.Unlock()

	return aof.file.Close()
}

func (aof *Aof) Write(value Value) error {
	aof.mutex.Lock()
	defer aof.mutex.Unlock()

	_, err := aof.file.Write(value.Serialize())
	if err != nil {
		return err
	}

	return nil
}

func (aof *Aof) Read(callback func(value Value)) error {
	aof.mutex.Lock()
	defer aof.mutex.Unlock()

	resp := NewRespReader(aof.file)

	for {
		value, err := resp.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		callback(value)
	}

}
