// ttfb.go: a little TCP server for experimenting with TTFB
// measurements.  This server pretends to be an HTTP server and
// replies to any request with a 200 OK.  It does this by sending a
// single byte containing the H of HTTP/1.1 200 OK and then waits 10
// seconds before sending the rest of the headers.
//
// Copyright (c) 2012 CloudFlare, Inc.

package main

import (
	"time"
	"fmt"
	"flag"
	"net"
	"bufio"
	"strings"
)

func main() {

	// Set -port=X on the command-line to override the default port of
	// 8888

	var port string
	flag.StringVar(&port, "port", "8888", "port to listen on")
	flag.Parse()

	// The HTTP status line, response body and response headers that
	// will be returned for any HTTP request

	status  := "HTTP/1.1 200 OK\r\n"
	body    := "Hello, World!\r\n"
	headers := "Content-Type: text/html\r\n" +
		fmt.Sprintf("Content-Length: %d\r\n", len(body)) +
		"Server: golang test server\r\n" +
		fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123)) +
		"Cache-Control: no-cache\r\n\r\n"

	// Listen for TCP connections on any interface on the port and
	// accept them as they come in.  To allow multiple connections
	// each connection is handled by a separate goroutine that handles
	// a single HTTP request and then terminates.

	var l net.Listener
	var err error
	if l, err = net.Listen("tcp", ":" + port); err != nil {
		fmt.Printf("Error listening on port %s: %s\n", port, err)
		return
	}

	for {
		var c net.Conn
		if c, err = l.Accept();  err != nil {
			fmt.Printf("Error accepting connection: %s\n", err)
			return
		}

		go func() {
			r := bufio.NewReader(c)
			
			// The actual HTTP request is read and completely ignored,
			// all this does is look for the blank line at the end of
			// the request headers.
			//
			// Note that this loop is deliberately ignoring any errors
			// or any truncated lines so it is not robust.  Throughout
			// the rest of this code errors on the network connection
			// are completely ignored.

			for {
				b, _, _ := r.ReadLine();
				if strings.TrimSpace(string(b)) == "" {
					break
				}
			}

			w := bufio.NewWriter(c)
				
			// Write out the first character of the HTTP status line
			// (which will be H) and flush it to make sure that it
			// gets sent.  Then wait 10 seconds before sending
			// anything else.
				
			w.WriteByte(status[0])
			w.Flush()
			time.Sleep(1e9 * 10)
			
			w.WriteString(status[1:])
			w.WriteString(headers)
			w.WriteString(body)
			w.Flush()
				
			c.Close()
		}()
	}
}
