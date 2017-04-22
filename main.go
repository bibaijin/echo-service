package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

const (
	// LogFlag 控制日志的前缀
	LogFlag = log.LstdFlags | log.Lmicroseconds | log.Lshortfile
	// MaxConnectionNum 表示最大连接数
	MaxConnectionNum = 10000
)

var (
	errLogger  = log.New(os.Stderr, "ERROR ", LogFlag)
	infoLogger = log.New(os.Stdout, "INFO ", LogFlag)
)

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM)

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("net.Listen failed, error: %s.", err)
	}

	infoLogger.Printf("net.Listen() %s...", ":8080")

	throttle := make(chan struct{}, MaxConnectionNum)
	var wg sync.WaitGroup
	running := true
	for running {
		conn, err := ln.Accept()
		if err != nil {
			errLogger.Printf("ln.Accept failed, error: %s.", err)
		} else {
			infoLogger.Printf("Accept a connection, RemoteAddr: %s.", conn.RemoteAddr())
			wg.Add(1)
			throttle <- struct{}{}

			go func(conn net.Conn) {
				defer wg.Done()
				handle(conn, quit, throttle)
			}(conn)
		}

		select {
		case signal := <-quit:
			infoLogger.Printf("Receive a signal: %d, and I will shutdown gracefully...", signal)
			running = false
		default:
			infoLogger.Printf("Ready for another connection.")
		}
	}

	wg.Wait()
	infoLogger.Print("Shutdown gracefully.")
}

func handle(conn io.ReadWriteCloser, quit chan os.Signal, throttle <-chan struct{}) {
	defer func() {
		if err := conn.Close(); err != nil {
			errLogger.Printf("conn.Close() failed, error: %s.", err)
		}
	}()

	buf := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	running := true
	for running {
		s, err := buf.ReadString('\n')
		if err != nil {
			errLogger.Printf("buf.ReadString() failed, error: %s.", err)
			running = false
		}

		infoLogger.Printf("Read a message: %s.", strings.TrimSpace(s))

		n, err := buf.WriteString(s)
		if err != nil {
			errLogger.Printf("buf.WriteString() failed, error: %s.", err)
			running = false
		}

		if err = buf.Flush(); err != nil {
			errLogger.Printf("buf.Flush() failed, error: %s.", err)
			running = false
		}

		infoLogger.Printf("Write a response: %s, length: %d bytes.", strings.TrimSpace(s), n)

		select {
		case signal := <-quit:
			infoLogger.Printf("Receive a signal: %d, and I will shutdown gracefully...", signal)
			running = false
			quit <- signal
		default:
			infoLogger.Printf("Ready for another message.")
		}
	}

	<-throttle
}
