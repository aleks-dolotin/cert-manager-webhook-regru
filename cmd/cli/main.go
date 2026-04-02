// CLI tool for managing TXT records in Reg.ru DNS.
//
// Usage:
//   REGRU_USERNAME=user REGRU_PASSWORD=pass go run ./cmd/cli create dolotin.ru _acme-test "hello"
//   REGRU_USERNAME=user REGRU_PASSWORD=pass go run ./cmd/cli delete dolotin.ru _acme-test "hello"
package main

import (
	"fmt"
	"os"

	"github.com/aleks-dolotin/cert-manager-webhook-regru/internal/regru"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	username := os.Getenv("REGRU_USERNAME")
	password := os.Getenv("REGRU_PASSWORD")
	if username == "" || password == "" {
		fmt.Fprintln(os.Stderr, "Error: REGRU_USERNAME and REGRU_PASSWORD must be set")
		os.Exit(1)
	}

	client := regru.NewClient(username, password)
	command := os.Args[1]

	switch command {
	case "create":
		if len(os.Args) < 5 {
			fmt.Fprintln(os.Stderr, "Usage: cli create <zone> <subdomain> <content>")
			os.Exit(1)
		}
		zone, subdomain, content := os.Args[2], os.Args[3], os.Args[4]
		fmt.Printf("Creating TXT: %s.%s = %q\n", subdomain, zone, content)
		if err := client.CreateTXT(zone, subdomain, content); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Created!")

	case "delete":
		if len(os.Args) < 5 {
			fmt.Fprintln(os.Stderr, "Usage: cli delete <zone> <subdomain> <content>")
			os.Exit(1)
		}
		zone, subdomain, content := os.Args[2], os.Args[3], os.Args[4]
		fmt.Printf("Deleting TXT: %s.%s = %q\n", subdomain, zone, content)
		if err := client.DeleteTXT(zone, subdomain, content); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Deleted!")

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  cli create <zone> <subdomain> <content>  — create TXT record")
	fmt.Fprintln(os.Stderr, "  cli delete <zone> <subdomain> <content>  — delete TXT record")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Example:")
	fmt.Fprintln(os.Stderr, "  REGRU_USERNAME=user REGRU_PASSWORD=pass go run ./cmd/cli create dolotin.ru _acme-test hello")
}
