package main

import (
	"os"

	"github.com/aleks-dolotin/cert-manager-webhook-regru/internal/solver"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
)

// GroupName is the API group name for this webhook solver.
// Must match the groupName in the ClusterIssuer solver config.
var GroupName = os.Getenv("GROUP_NAME")

func main() {
	if GroupName == "" {
		GroupName = "acme.dolotin.ru"
	}

	cmd.RunWebhookServer(GroupName, solver.New())
}
