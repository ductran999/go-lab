// Command drip Slowloris-probes a server: one header line per interval,
// never completing the block. Prints when (and how) the server cuts
// the connection — or that it is still waiting politely.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// run drips until the server cuts us or we run out of lines.
func run() error {
	addr := flag.String("addr", "localhost:8121", "server to drip")
	every := flag.Int("every", 200, "ms between header lines")
	lines := flag.Int("lines", 100, "max header lines before giving up")

	flag.Parse()

	conn, err := net.DialTimeout("tcp", *addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	defer func() {
		_ = conn.Close()
	}()

	_, err = conn.Write([]byte("GET /fast HTTP/1.1\r\nHost: x\r\n"))
	if err != nil {
		return fmt.Errorf("first write: %w", err)
	}

	start := time.Now()

	for i := range *lines {
		time.Sleep(time.Duration(*every) * time.Millisecond)

		header := fmt.Sprintf("X-Drip-%d: y\r\n", i)
		log.Printf("drip %d: %s", i, header)

		// Example of how to end the header block after 5 lines, uncomment to test
		// if i == 5 {
		// 	_, err = fmt.Fprintf(conn, "\r\n")
		// 	if err != nil {
		// 		fmt.Println("write end header got error", err)
		// 	}
		// 	fmt.Println("drip ended after 5 lines, server should respond now")
		// 	return nil
		// }

		_, err = fmt.Fprint(conn, header)
		if err != nil {
			fmt.Printf("server cut after %d lines / %v: %v\n", i, time.Since(start).Round(time.Millisecond), err)

			return nil
		}
	}

	fmt.Printf("server still waiting after %d lines / %v\n", *lines, time.Since(start).Round(time.Millisecond))

	return nil
}
