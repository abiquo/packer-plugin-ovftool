package main

import (
	"fmt"
	ovftool "github.com/chirauki/packer-post-processor-ovftool/ovftool"
	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"os"
)

func main() {
	//server, err := plugin.Server()
	//if err != nil {
	//	panic(err)
	//}
	//server.RegisterPostProcessor(new(ovftool.PostProcessor))
	//server.Serve()
	pps := plugin.NewSet()
	pps.RegisterPostProcessor("ovftool", new(ovftool.PostProcessor))

	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
