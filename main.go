package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/0xBF/packer-post-processor-ovftool/ovftool"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterPostProcessor(new(ovftool.OVFtoolPostProcessor))
	server.Serve()
}
