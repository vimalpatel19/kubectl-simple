package main

import (
	"github.com/vimalpatel19/kubectl-simple/cmd/plugin/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // required for GKE
)

func main() {
	cli.Execute()
}
