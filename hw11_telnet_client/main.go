package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage of %s:
  --timeout duration
    connection timeout (default 10s)

  host
    ip address or domain name to connect to (required)

  port
    port to connect to (required)

Example: %s --timeout=5s hostname 8080
`, os.Args[0], os.Args[0])
	}
}

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	// Must be 2 args.
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	// Get host and port
	host, port := flag.Arg(0), flag.Arg(1)

	// Proper concatenation for IPv6.
	hostport := net.JoinHostPort(host, port)
	telnet := NewTelnetClient(hostport, *timeout, os.Stdin, os.Stdout)

	fmt.Fprintf(os.Stderr, "...Trying %s\n", hostport)

	if err := telnet.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "...Connection to %s failed: %v\n", hostport, err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "...Connected to %s\n", hostport)

	// Handle any communication error here.
	// Exit status value is based on non-nil error.
	var err error

	// Terminate func: close connection, set exit status value on error.
	defer func() {
		if cerr := telnet.Close(); cerr != nil {
			err = cerr
			fmt.Fprintf(os.Stderr, "...Close connection error: %v\n", cerr)
		}

		// There was a communication error.
		if err != nil {
			os.Exit(1)
		}
	}()

	// Handle Ctrl-C
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	select {
	case err = <-receive(telnet):
		if err != nil {
			fmt.Fprintf(os.Stderr, "...Connection read error: %v\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "...Connection was closed by peer")
		}
	case err = <-send(telnet):
		if err != nil {
			fmt.Fprintf(os.Stderr, "...Connection write error: %v\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "...EOF")
		}
	case <-signals:
		// Second Ctrl-C forces to terminate app
		signal.Reset()
		fmt.Fprintln(os.Stderr, "...Termitate connection")
	}
}

// receive does telnet.Receive and returns error-channel.
// Any occurred error will be send to the channel.
// Nil error in channel means the remote peer closed the connection.
func receive(telnet TelnetClient) <-chan error {
	c := make(chan error, 1)
	go func() {
		defer close(c)

		err := telnet.Receive()
		c <- err
	}()
	return c
}

// receive does telnet.Send and returns error-channel.
// Any occurred error will be send to the channel.
// Nil error in channel means the communication is finished because of EOF.
func send(telnet TelnetClient) <-chan error {
	c := make(chan error, 1)

	go func() {
		defer close(c)

		err := telnet.Send()
		c <- err
	}()

	return c
}
