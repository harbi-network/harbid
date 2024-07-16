# harbiminer

Harbiminer is a CPU-based miner for harbid

## Requirements

Go 1.19 or later.

## Installation

#### Build from Source

- Install Go according to the installation instructions here:
  http://golang.org/doc/install

- Ensure Go was installed properly and is a supported version:

```bash
$ go version
```

- Run the following commands to obtain and install harbid including all dependencies:

```bash
$ git clone https://github.com/harbi-network/harbid
$ cd harbid/cmd/harbiminer
$ go install .
```

- harbiminer should now be installed in `$(go env GOPATH)/bin`. If you did
  not already add the bin directory to your system path during Go installation,
  you are encouraged to do so now.
  
## Usage

The full harbiminer configuration options can be seen with:

```bash
$ harbiminer --help
```

But the minimum configuration needed to run it is:
```bash
$ harbiminer --miningaddr=<YOUR_MINING_ADDRESS>
```