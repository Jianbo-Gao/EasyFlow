## EasyFlow

EasyFlow is a prototype of overflow detection tool for Ethereum smart contracts. 

This tool is developed based on the official `evm` tool in [**go-ethereum** ](https://github.com/ethereum/go-ethereum). 

A brief introduction video has been uploaded to [Youtube](https://youtu.be/J6EKP1crtwI). 

## Building the source

### Installation instructions for Ubuntu 16.04

(Typically these commands must be run as root or through `sudo`.) 

Install latest distribution of Go(v1.10.3):

```bash
wget https://dl.google.com/go/go1.10.3.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.10.3.linux-amd64.tar.gz
```

Add `/usr/local/go/bin` to the `PATH` environment variable. You can do this by adding this line to `/etc/profile`:

```bash
export PATH=$PATH:/usr/local/go/bin
```

Apply the changes in `/etc/profile` immediately:

```bash
source /etc/profile
```

Install C compilers: 

```bash
apt-get install -y build-essential
```

Clone the repository to a directory of your choosing build:

```bash
git clone git@github.com:Jianbo-Gao/EasyFlow.git
```

Finally, build EasyFlow core module using the following command.

```bash
cd EasyFlow
make evm
```

You can now run `python run.py` in `taint_scripts` to use EasyFlow.


## Using the tool
### Using EasyFlow

Examples can be accessed in `taint_scripts/cmd.sh`

### Using Modified EVM Tool (Core module of EasyFlow)

Examples can be accessed in `taint_contracts/cmd.sh`

