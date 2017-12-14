# Packer OVFtool post processor

Just another OVF tool post processor.

## Requirements

- VMware's [OVF Tool](https://www.vmware.com/support/developer/ovf/).
- OpenSSL command.

## Usage

Add the post-processor to your packer template:

```
    {
      "type": "ovftool",
      "target_format": "ova",
      "appliance_name": "{{user `vm_name`}}",
      "output_dir": "{{user `output_dir`}}",
      "keep_ovf": "{{user `keep_ovf`}}"
    },
```

Available configuration options:

| Key                   | Desc  | Default  |
|-----------------------|---|---|
| `ovftool_path`        | Path to the `ovftool` binary if it cannot be found in the path. | `ovftool` |
| `output_dir`          | The directory where to save the resulting appliance/s. | `output/packer_{{ .BuildName }}_{{ .Provider }}_ovftool` |
| `keep_input_artifact` | Wether or not to keep the input artifact | false |
| `target_format`       | Either `ovf` or `ova`. Use `keep_ovf` and `ova` if you want both OVF and OVA | `ovf` |
| `keep_ovf`            | If `target_format` is OVA the initial OVF export will be deleted unless this param is true | false |
| `appliance_name`      | The name to give to the resulting appliance |  |

## Installation

You can grab pre-built binaries from:

https://github.com/chirauki/packer-post-processor-ovftool/releases/latest

And add the post-processor to ~/.packerconfig:

```
{
  "post-processors": {
    "ovftool": "packer-post-processor-ovftool"
  }
}
```

Or if you want to build it yourself, run:

```
$ go get github.com/chirauki/packer-post-processor-ovftool
$ go install github.com/chirauki/packer-post-processor-ovftool
```

And add the post-processor to ~/.packerconfig.

# License and Authors

* Author:: Marc Cirauqui (marc.cirauqui@abiquo.com)

Copyright:: 2014, Abiquo

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
