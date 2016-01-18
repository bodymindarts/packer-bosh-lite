package provisioner

import (
	"fmt"
	"os"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Stemcell        string
	StemcellVersion string `mapstructure:"stemcell_version"`

	Release        string
	ReleaseVersion string `mapstructure:"release_version"`

	Manifest string `mapstructure:"deployment_manifest"`

	ctx interpolate.Context
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) Prepare(raws ...interface{}) error {

	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {

	p.uploadStemcell(ui, comm)
	p.uploadBoshRelease(ui, comm)
	p.uploadDeploymentManifest(ui, comm)
	p.deploy(ui, comm)

	return nil
}

func (p *Provisioner) Cancel() {
	os.Exit(0)
}

func (p *Provisioner) uploadStemcell(ui packer.Ui, comm packer.Communicator) error {
	ui.Say(fmt.Sprintf("Uploading stemcell %s", p.config.Stemcell))

	cmd := fmt.Sprintf("bosh upload stemcell --skip-if-exists https://bosh.io/d/stemcells/%s", p.config.Stemcell)
	if p.config.StemcellVersion != "" {
		cmd += fmt.Sprintf("?v=%s", p.config.StemcellVersion)
	}

	return runCmd(cmd, ui, comm)
}

func (p *Provisioner) uploadBoshRelease(ui packer.Ui, comm packer.Communicator) error {
	ui.Say(fmt.Sprintf("Uploading release %s", p.config.Release))

	cmd := fmt.Sprintf("bosh upload release --skip-if-exists https://bosh.io/d/github.com/%s", p.config.Release)
	if p.config.ReleaseVersion != "" {
		cmd += fmt.Sprintf("?v=%s", p.config.ReleaseVersion)
	}

	return runCmd(cmd, ui, comm)
}

func (p *Provisioner) uploadDeploymentManifest(ui packer.Ui, comm packer.Communicator) error {
	ui.Say(fmt.Sprintf("Uploading manifest: %s", p.config.Manifest))

	err := runCmd("mkdir ~/deployments", ui, comm)
	if err != nil {
		return err
	}

	f, err := os.Open(p.config.Manifest)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	err = comm.Upload(fmt.Sprintf("~/deployments/%s", p.config.Manifest), f, &fi)
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("sed -i \"s/director_uuid: .*/director_uuid: $(bosh status --uuid)/\" ~/deployments/%s", p.config.Manifest)
	return runCmd(cmd, ui, comm)
}

func (p *Provisioner) deploy(ui packer.Ui, comm packer.Communicator) error {
	ui.Say("Deploying")

	cmd := fmt.Sprintf("bosh deployment ~/deployments/%s", p.config.Manifest)
	err := runCmd(cmd, ui, comm)
	if err != nil {
		return err
	}

	cmd = "sudo mkdir -p /vagrant/tmp/compiled_package_cache && sudo chmod -R a+rw /vagrant"
	err = runCmd(cmd, ui, comm)
	if err != nil {
		return err
	}

	err = runCmd("bosh -n deploy", ui, comm)
	if err != nil {
		return err
	}

	return runCmd("sudo rm -rf /vagrant", ui, comm)
}

func runCmd(cmd string, ui packer.Ui, comm packer.Communicator) error {
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
