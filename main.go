package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	fmt.Println("gangway: Docker API shim for Apple containers")
	fmt.Println()
	fmt.Println("usage: gangway [flags]")
	fmt.Println()
	fmt.Println("flags:")
	fmt.Println("  --socket <path>         unix socket path (default: /tmp/gangway.sock)")
	fmt.Println("  --container-bin <path>  Apple container CLI path (default: container)")
	fmt.Println("  --log-level <level>     log verbosity (default: info)")
	return nil
}
