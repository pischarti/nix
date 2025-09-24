package main

import (
	"gofr.dev/pkg/gofr"

	"github.com/pischarti/nix/go/pkg/container"
)

func main() {
	app := gofr.NewCMD()

	app.SubCommand("images", container.ImagesHandler,
		gofr.AddDescription("List container images running in the cluster"),
		gofr.AddHelp("Usage: kube images [--namespace NAMESPACE | --all-namespaces] [--by-pod] [--table] [--style STYLE] [--sort SORT]"),
	)

	app.Run()
}
