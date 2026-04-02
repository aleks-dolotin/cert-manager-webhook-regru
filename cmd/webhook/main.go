package main

import (
	"fmt"
	"os"

	"github.com/aleks-dolotin/cert-manager-webhook-regru/internal/solver"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
)

// GroupName is the API group name for this webhook solver.
// Must match the groupName in the ClusterIssuer solver config.
var GroupName = os.Getenv("GROUP_NAME")

// Version info injected at build time via ldflags.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("cert-manager-webhook-regru %s (commit: %s, built: %s)\n", Version, Commit, BuildDate)
		return
	}

	if GroupName == "" {
		GroupName = "acme.dolotin.ru"
	}

	cmd.RunWebhookServer(GroupName, solver.New())
}
