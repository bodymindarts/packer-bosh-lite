package provisioner

import (
	"fmt"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
)

type Config struct {
	metadata *mapstructure.Metadata

	common.PackerConfig `mapstructure:",squash"`

	Release string `mapstructure:"release"`

	Version string `mapstructure:"release_version"`
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        false,
		InterpolateContext: nil,
		InterpolateFilter:  nil},
		raws...)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {

	p.uploadBoshRelease(ui, comm)
	return nil
}

func (p *Provisioner) Cancel() {
	os.Exit(0)
}

func (p *Provisioner) uploadBoshRelease(ui packer.Ui, comm packer.Communicator) error {
	ui.Say(fmt.Sprintf("Uploading release %s", p.config.Release))

	cmd := fmt.Sprintf("bosh upload release https://bosh.io/d/github.com/%s", p.config.Release)
	if p.config.Version != "" {
		cmd += fmt.Sprintf("?v=%s", p.config.Version)
	}

	remoteCmd := &packer.RemoteCmd{Command: cmd}
	err := remoteCmd.StartWithUi(comm, ui)
	if err != nil {
		return fmt.Errorf("Starting command: %s", err)
	}

	if remoteCmd.ExitStatus != 0 {
		return fmt.Errorf("Non-zero exit status: %d", remoteCmd.ExitStatus)
	}

	return nil
}
