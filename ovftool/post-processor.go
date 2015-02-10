package ovftool

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	//vmwiso "github.com/mitchellh/packer/builder/vmware/iso"
)


type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	OVFtoolPath string `mapstructure:"ovftool_path"`
	OutputDir   string `mapstructure:"output_dir"`

	Host     string `mapstructure:"host"`
	SshPort  int    `mapstructure:"ssh_port"`
	ViPort   int    `mapstructure:"vi_port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	VMName   string `mapstructure:"vm_name"`

	tpl *packer.ConfigTemplate
}

type OVFtoolPostProcessor struct {
	cfg Config

	comm packer.Communicator
	vmId string
	vmxPath string
}

type OutputPathTemplate struct {
	ArtifactId string
	BuildName  string
	Provider   string
}

func (p *OVFtoolPostProcessor) Configure(raws ...interface{}) error {
	_, err := common.DecodeConfig(&p.cfg, raws...)
	if err != nil {
		return err
	}

	p.cfg.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}

	p.cfg.tpl.UserVars = p.cfg.PackerUserVars

	if p.cfg.OVFtoolPath == "" {
		p.cfg.OVFtoolPath = "ovftool"
	}

	if p.cfg.OutputDir == "" {
		p.cfg.OutputDir = "output/packer_{{ .BuildName }}_{{ .Provider }}_ovftool"
	}

	if p.cfg.SshPort == 0 {
		p.cfg.SshPort = 22
	}

	if p.cfg.ViPort == 0 {
		p.cfg.ViPort = 22
	}

	if p.cfg.Username == "" {
		p.cfg.Username = "root"
	}

	// Accumulate any errors
	errs := new(packer.MultiError)

	templates := map[string]*string {
		"ovftool_path": &p.cfg.OVFtoolPath,
		"output_dir": &p.cfg.OutputDir,
		"host": &p.cfg.Host,
//		"port": &p.cfg.Port,
		"username": &p.cfg.Username,
		"vm_name": &p.cfg.VMName,
		"password": &p.cfg.Password,
	}

	for key, ptr := range templates {
		*ptr, err = p.cfg.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", key, err))
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	if p.cfg.Host == "" {
		return fmt.Errorf("ovftool post processor: host parameter is required")
	}

	if p.cfg.Password == "" {
		return fmt.Errorf("ovftool post processor: password parameter is required")
	}

	if p.cfg.VMName == "" {
		return fmt.Errorf("ovftool post processor: vm_name parameter is required")
	}


	return nil
}

func (p *OVFtoolPostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	if artifact.BuilderId() != "mitchellh.vmware-esx" {
		return nil, false, fmt.Errorf("ovftool post-processor can only be used on VMware ESX builds: %s", artifact.BuilderId())
	}

	for _, f := range artifact.Files() {
		if strings.HasSuffix(f, ".vmx") {
			p.vmxPath = f
		}
	}

	if p.vmxPath == "" {
		return nil, false, fmt.Errorf("No .vmx file in artifact")
	}


	ui.Say( "Cleaning output directory...")

	err := os.RemoveAll( p.cfg.OutputDir)
	if err != nil {
		return nil, false, err
	}


	ui.Say( "Registering VM...")

	err = p.connect()
	if err != nil {
		return nil, false, err
	}

	err = p.Register()
	if err != nil {
		return nil, false, err
	}
	defer p.Unregister()


	ui.Say( "Exporting VM...")

	var stdout, stderr bytes.Buffer
	source := fmt.Sprintf( "vi://%s:%s@%s:%s/%s", p.cfg.Username, p.cfg.Password, p.cfg.Host, p.cfg.ViPort, p.cfg.VMName)

	cmd := exec.Command( p.cfg.OVFtoolPath, "--noSSLVerify", source, p.cfg.OutputDir)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return nil, false, fmt.Errorf("Unable to execute ovftool:\n== STDOUT ==\n%s== STDERR ==\n%s", stdout.String(), stderr.String())
	}

	vmdir := fmt.Sprintf( "%s/%s", p.cfg.OutputDir, p.cfg.VMName)
	files, _ := ioutil.ReadDir( vmdir)
	for _, f := range files {
		source := fmt.Sprintf( "%s/%s", vmdir, f.Name())
		dest := fmt.Sprintf( "%s/%s", p.cfg.OutputDir, f.Name())
		os.Rename( source, dest)
	}
	os.Remove( vmdir)

	return &Artifact{ dir: p.cfg.OutputDir}, false, nil
}
