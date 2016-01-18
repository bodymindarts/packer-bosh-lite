package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/bodymindarts/packer-bosh-lite/provisioner"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}

	server.RegisterProvisioner(new(provisioner.Provisioner))

	server.Serve()
}
