package ovftool

import (
	"fmt"
	"os"
)

const BuilderId = "packer.post-processor.ovftool"

type Artifact struct {
	dir string
}

func (a *Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return []string{a.dir}
}

func (*Artifact) Id() string {
	return "OVF"
}

func (a *Artifact) String() string {
	return fmt.Sprintf("OVF template in directory: %s", a.dir)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return os.RemoveAll(a.dir)
}
