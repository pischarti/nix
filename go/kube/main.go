package main

import (
	"gofr.dev/pkg/gofr"

	"github.com/pischarti/nix/pkg/container"
)

func main() {
	app := gofr.NewCMD()

	app.SubCommand("images", container.ImagesHandler,
		gofr.AddDescription("List container images running in the cluster"),
		gofr.AddHelp("Usage: kube images [--namespace NAMESPACE | --all-namespaces] [--by-pod] [--table] [--style STYLE] [--sort SORT]"),
	)

	app.SubCommand("services", container.ServicesHandler,
		gofr.AddDescription("List Kubernetes services with annotations matching specified criteria"),
		gofr.AddHelp("Usage: kube services [--namespace NAMESPACE | --all-namespaces] [--table] [--style STYLE] [--sort SORT] [--annotation-value VALUE]"),
	)

	app.Run()
}
