package main

import (
	ovftool "github.com/chirauki/packer-post-processor-ovftool/ovftool"
	"github.com/hashicorp/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterPostProcessor(new(ovftool.PostProcessor))
	server.Serve()
}
