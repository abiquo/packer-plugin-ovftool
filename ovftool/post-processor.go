package ovftool

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

const (
	Separator = string(os.PathSeparator)
)

type PostProcessor struct {
	config Config
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	OVFtoolPath string `mapstructure:"ovftool_path"`
	OutputDir   string `mapstructure:"output_dir"`
	KeepSource  bool   `mapstructure:"keep_input_artifact"`

	TargetFormat  string `mapstructure:"target_format"`
	KeepOvf       bool   `mapstructure:"keep_ovf"`
	ApplianceName string `mapstructure:"appliance_name"`

	ctx interpolate.Context
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate: true,
	}, raws...)
	if err != nil {
		return err
	}

	if p.config.ApplianceName == "" {
		errorMsg := "Appliance name is mandatory!"
		return errors.New(errorMsg)
	}

	if p.config.OVFtoolPath == "" {
		p.config.OVFtoolPath = "ovftool"
	}

	if p.config.OutputDir == "" {
		p.config.OutputDir = "output" + Separator + "packer_{{ .BuildName }}_{{ .Provider }}_ovftool"
	}

	if p.config.TargetFormat == "" {
		p.config.TargetFormat = "ovf"
	}

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	keep := p.config.KeepSource

	if artifact.BuilderId() != "mitchellh.vmware" {
		return nil, keep, fmt.Errorf("ovftool post-processor can only be used on VMware builds: %s", artifact.BuilderId())
	}

	log.Println("Looking for the VMX file...")
	var vmxPath string
	for _, f := range artifact.Files() {
		if strings.HasSuffix(f, ".vmx") {
			vmxPath = f
		}
	}

	if vmxPath == "" {
		return nil, keep, fmt.Errorf("No .vmx file in artifact")
	}
	log.Printf("VMX file is at %s", vmxPath)
	log.Printf("Running OVFtool...")

	source := path.Clean(vmxPath)
	dest := path.Clean(p.config.OutputDir + Separator + p.config.ApplianceName + Separator + p.config.ApplianceName + ".ovf")
	log.Printf("Source is: %s", source)
	log.Printf("Target is: %s", dest)

	// MAKE sure dest path exists or OVFTool will
	// do funky things.
	os.Mkdir(path.Dir(dest), os.ModePerm)
	cmdname := p.config.OVFtoolPath
	cmdargs := []string{"-o", "-n=" + p.config.ApplianceName, source, dest}
	log.Printf("OVFTool command: %s %s", cmdname, cmdargs)
	out, err := exec.Command(cmdname, cmdargs...).Output()
	if err != nil {
		log.Printf("Error executing OVFTool! %s", err)
		log.Printf("OVFTool: %s", out)
		return nil, keep, err
	}
	log.Printf("OVFTOOL: %s", out)

	// Modify ID to name in VirtualSystem ID
	log.Printf("Setting the VM name in VirtualSystem ID.")
	var xmlLines []string
	xmlFile, err := ioutil.ReadFile(dest)
	if err != nil {
		log.Println("Error opening OVF file:", err)
		return nil, keep, err
	}

	productXml := ""
	iconFileRef := ""
	if _, info := p.config.PackerUserVars["info"]; info {
		if _, prod := p.config.PackerUserVars["product"]; prod {
			productXml = fmt.Sprintf("<ProductSection><Info>%s</Info><Product>%s</Product>",
				p.config.PackerUserVars["info"],
				p.config.PackerUserVars["product"])
			if _, ic := p.config.PackerUserVars["icon"]; ic {
				iconFileRef = fmt.Sprintf("<File ovf:href=\"%s\" ovf:id=\"icon\"/>", p.config.PackerUserVars["icon"])
				productXml = productXml +
					"<Icon ovf:fileRef=\"icon\" ovf:mimeType=\"image/jpeg\" ovf:width=\"32\" ovf:height=\"32\"/>"
			}
			if _, user := p.config.PackerUserVars["ssh_username"]; user {
				productXml = productXml +
					fmt.Sprintf("<Property ovf:key=\"user\" ovf:type=\"string\" ovf:value=\"%s\">",
						p.config.PackerUserVars["ssh_username"]) +
					"<Description>Default login username</Description></Property>"
			}
			if _, pass := p.config.PackerUserVars["ssh_password"]; pass {
				productXml = productXml +
					fmt.Sprintf("<Property ovf:key=\"password\" ovf:type=\"string\" ovf:value=\"%s\">",
						p.config.PackerUserVars["ssh_password"]) +
					"<Description>Default login password</Description></Property>"
			}
			productXml = productXml + "</ProductSection>"
		}
	}

	lines := strings.Split(string(xmlFile), "\n")
	for _, line := range lines {
		if strings.Contains(line, "<File ovf:href") {
			xmlLines = append(xmlLines, line)
			if iconFileRef != "" {
				xmlLines = append(xmlLines, iconFileRef)
			}
		} else if strings.Contains(line, "VirtualSystem ovf:id=\"vm\"") {
			sline := strings.Replace(line, "VirtualSystem id=\"vm\"", "VirtualSystem id=\""+p.config.ApplianceName+"\"", -1)
			log.Printf("VS ID changed: %s", sline)
			xmlLines = append(xmlLines, sline)
		} else if strings.Contains(line, "</OperatingSystemSection>") {
			xmlLines = append(xmlLines, line)
			if productXml != "" {
				xmlLines = append(xmlLines, productXml)
			}
		} else {
			xmlLines = append(xmlLines, line)
		}
	}
	log.Printf("xmlLines: %q", strings.Join(xmlLines, ""))

	writer, err := os.Create(dest)
	if err != nil {
		return nil, keep, err
	}
	defer writer.Close()
	for _, line := range xmlLines {
		writer.WriteString(line + "\n")
	}
	writer.Sync()

	// Now rebuild the manifest
	log.Printf("Regenerating OVF manifest file.")
	cmdname = "openssl"
	cmdargs = []string{"sha1"}
	ovfFilesPath := path.Clean(p.config.OutputDir + Separator + p.config.ApplianceName)
	log.Printf("Scanning files in %s", ovfFilesPath)
	var mfFile string
	files, _ := ioutil.ReadDir(ovfFilesPath)
	for _, f := range files {
		log.Printf("Found file %s", f.Name())
		if strings.HasSuffix(f.Name(), "mf") {
			mfFile = f.Name()
		} else {
			cmdargs = append(cmdargs, f.Name())
		}
	}
	cmd := exec.Command(cmdname, cmdargs...)
	cmd.Dir = ovfFilesPath
	log.Printf("OpenSSL command: %s %s", cmdname, cmdargs)
	log.Printf("Ruunning OpenSSL from %s", cmd.Dir)
	output, _ := cmd.CombinedOutput()

	log.Printf("OpenSSL output: %s", output)

	writer, err = os.Create(ovfFilesPath + Separator + mfFile)
	if err != nil {
		return nil, keep, err
	}
	defer writer.Close()
	writer.Write(output)
	writer.Sync()

	if p.config.TargetFormat == "ova" {
		// Create OVA from resulting OVF
		source = dest
		dest := path.Clean(p.config.OutputDir + Separator + p.config.ApplianceName + ".ova")

		log.Printf("OVA Source is: %s", source)
		log.Printf("OVA Target is: %s", dest)

		cmdname := p.config.OVFtoolPath
		cmdargs := []string{"-o", source, dest}
		log.Printf("OVA OVFTool command: %s %s", cmdname, cmdargs)
		out, err := exec.Command(cmdname, cmdargs...).Output()
		if err != nil {
			log.Printf("Error executing OVA OVFTool! %s", err)
			log.Printf("OVA OVFTool: %s", out)
			return nil, keep, err
		}
		log.Printf("OVA OVFTOOL: %s", out)

		if !p.config.KeepOvf {
			os.RemoveAll(path.Dir(source))
		}
	}

	return &Artifact{dir: p.config.OutputDir + Separator + p.config.ApplianceName}, keep, nil
}
