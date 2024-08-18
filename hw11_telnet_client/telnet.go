package main

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

var (
	ErrAlreadyConnected = errors.New("client is already connected")
	ErrNoConnection     = errors.New("client has no connection")
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

// telnetClient implements TelnetClient.
type telnetClient struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer

	conn net.Conn
	mx   sync.RWMutex // mx is a mutex-guard for the conn
}

// Connect tries to make a connection to telnetClient address.
//
// Returns ErrAlreadyConnected if already connected or connection error.
func (c *telnetClient) Connect() error {
	c.mx.Lock()
	defer c.mx.Unlock()

	if c.conn != nil {
		return ErrAlreadyConnected
	}

	conn, err := net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return err
	}

	c.conn = conn

	return nil
}

// Close tries to close an opened connection.
//
// Returns ErrNoConnection if not connected, or error occurred on connection closing.
func (c *telnetClient) Close() error {
	c.mx.Lock()
	defer c.mx.Unlock()

	if c.conn == nil {
		return ErrNoConnection
	}

	err := c.conn.Close()
	c.conn = nil
	if err != nil {
		return err
	}

	return nil
}

// Send copies data from the telnetClient in to the connection.
//
// Returns ErrNoConnection  if not connected, or error occurred on coping.
func (c *telnetClient) Send() error {
	c.mx.RLock()
	conn := c.conn
	c.mx.RUnlock()

	if conn == nil {
		return ErrNoConnection
	}

	_, err := io.Copy(conn, c.in)

	return err
}

// Receive copies data from the connection to telnetClient out.
//
// Returns ErrNoConnection  if not connected, or error occurred on coping.
func (c *telnetClient) Receive() error {
	c.mx.RLock()
	conn := c.conn
	c.mx.RUnlock()

	if conn == nil {
		return ErrNoConnection
	}

	_, err := io.Copy(c.out, conn)

	return err
}
