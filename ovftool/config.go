package ovftool

import (
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

//go:generate packer-sdc mapstructure-to-hcl2 -type Config
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
