package provisioner

import (
	"os"

	"github.com/mitchellh/packer/packer"
)

type Provisioner struct {
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
    ui.Say("hello this is the packer-bosh-lite provisioner")
	return nil
}

func (p *Provisioner) Cancel() {
	os.Exit(0)
}
