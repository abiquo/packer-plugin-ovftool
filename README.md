Packer shell post-processor
=============================

Run shell scripts for post-process images

Usage
-----
Add the post-processor to your packer template:

    {
        "post-processors": [
          {
            "type": "ovftool",
            "host": "esxi01.local",
            "password": "top secret",
            "vm_name": "test"
          }
        ]
    }

Available configuration options:

  * ovftool_path
    Path to ovftool binary.

  * output_dir
    where to store OVF template files.

  * host
    Address of ESXi server. Should be accessable by ssh and by vi://

  * port
    Port for SSH access.

  * username
    Username for login by ssh and by vi://

  * pasword
    Password for login by ssh and by vi://

  * vm_name
    VM name. Template file names relate from this option.


Installation
------------
Run:

    $ go get github.com/0xBF/packer-post-processor-ovftool
    $ go install github.com/0xBF/packer-post-processor-ovftool

Add the post-processor to ~/.packerconfig:

    {
      "post-processors": {
        "ovftool": "packer-post-processor-ovftool"
      }
    }

