package main

import (
	"github.com/bodymindarts/packer-bosh-lite/provisioner"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}

	server.RegisterProvisioner(new(provisioner.Provisioner))

	server.Serve()
}
