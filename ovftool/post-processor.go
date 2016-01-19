package ovftool

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
	//vmwiso "github.com/mitchellh/packer/builder/vmware/iso"
)


type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	OVFtoolPath string `mapstructure:"ovftool_path"`
	OutputDir   string `mapstructure:"output_dir"`

	Host     string `mapstructure:"host"`
	SshPort  int    `mapstructure:"ssh_port"`
	//ViPort   int    `mapstructure:"vi_port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	VMName   string `mapstructure:"vm_name"`

	ctx interpolate.Context
}

type OVFtoolPostProcessor struct {
	cfg Config

	comm packer.Communicator
	vmId string
	vmxPath string
}

func (p *OVFtoolPostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.cfg, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if p.cfg.OVFtoolPath == "" {
		p.cfg.OVFtoolPath = "ovftool"
	}

	if p.cfg.OutputDir == "" {
		p.cfg.OutputDir = "output/packer_{{ .BuildName }}_{{ .Provider }}_ovftool"
	}

	if p.cfg.SshPort == 0 {
		p.cfg.SshPort = 22
	}

	//if p.cfg.ViPort == 0 {
	//	p.cfg.ViPort = 22
	//}

	if p.cfg.Username == "" {
		p.cfg.Username = "root"
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
	//source := fmt.Sprintf( "vi://%s:%s@%s:%d/%s", p.cfg.Username, p.cfg.Password, p.cfg.Host, p.cfg.ViPort, p.cfg.VMName)
	source := fmt.Sprintf( "vi://%s:%s@%s/%s", p.cfg.Username, p.cfg.Password, p.cfg.Host, p.cfg.VMName)

	cmd := exec.Command( p.cfg.OVFtoolPath, "--noSSLVerify", source, p.cfg.OutputDir)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
	  p.Unregister()
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
