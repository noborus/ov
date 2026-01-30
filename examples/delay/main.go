package main

import (
	"flag"
	"io"
	"os"
	"time"
)

func main() {
	// Accept delay in milliseconds (default 0)
	ms := flag.Int("ms", 0, "delay before forwarding EOF (in milliseconds)")
	flag.Parse()

	// Copy stdin to stdout directly
	_, err := io.Copy(os.Stdout, os.Stdin)
	if err != nil {
		os.Exit(1)
	}

	// When EOF is reached, wait
	time.Sleep(time.Duration(*ms) * time.Millisecond)
}
