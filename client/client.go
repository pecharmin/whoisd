// Copyright 2016 Openprovider Authors. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package client

import (
	"bytes"
	"log"
	"net"
	"os"

	"github.com/pecharmin/whoisd/storage"
	"golang.org/x/net/idna"
)

const (
	queryBufferSize = 256
)

// Record - standard record (struct) for client package
type Record struct {
	Conn  net.Conn
	Query []byte
}

// simplest logger, which initialized during starts of the application
var (
	stdlog = log.New(os.Stdout, "[CLIENT]: ", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "[CLIENT:ERROR]: ", log.Ldate|log.Ltime|log.Lshortfile)
)

// HandleClient - Sends a client data into the channel
func (client *Record) HandleClient(channel chan<- Record) {
	defer func() {
		if recovery := recover(); recovery != nil {
			errlog.Println("Recovered in HandleClient:", recovery)
			channel <- *client
		}
	}()
	buffer := make([]byte, queryBufferSize)
	numBytes, err := client.Conn.Read(buffer)
	if numBytes == 0 || err != nil {
		return
	}
	client.Query = bytes.ToLower(bytes.Trim(buffer, "\u0000\u000a\u000d"))
	channel <- *client
}

// ProcessClient - Asynchronous a client handling
func ProcessClient(channel <-chan Record, repository *storage.Record) {
	message := Record{}
	defer func() {
		if recovery := recover(); recovery != nil {
			errlog.Println("Recovered in ProcessClient:", recovery)
			if message.Conn != nil {
				message.Conn.Close()
			}
		}
	}()
	for {
		message = <-channel
		query, err := idna.ToASCII(string(message.Query))
		if err != nil {
			query = string(message.Query)
		}
		data, ok := repository.Search(query)
		message.Conn.Write([]byte(data))
		stdlog.Println(message.Conn.RemoteAddr().String(), query, ok)
		message.Conn.Close()
	}
}
