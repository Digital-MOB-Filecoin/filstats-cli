# filstats-cli

filstats-cli is a tool that connects to a local Filecoin node, extracts data from it and sends it to the filstats-server which centralizes data from multiple nodes to build a high level overview of the network.

## Currently supported node software
- [lotus](https://docs.filecoin.io/get-started/lotus/)

## Installation
### Build from source
**Prerequisites**
- a working Golang environment (tested with go v1.15)
    - requires go modules (>=go v1.11)
    
**Clone the repo**
```shell script
git clone git@github.com:Digital-MOB-Filecoin/filstats-cli.git
cd filstats-cli
```

**Build the executable**
```shell script
make
```

**Copy the sample config and do the necessary adjustments**
```shell script
cp config-sample.yml config.yml
```

**Start the server**
```shell script
./filstats-cli run
```

### Run via docker
```shell script
docker run -d \
  --restart always \
  --network="host" \
  --name filstats-client \
  -v ~/.filstats:/data \
  {TODO:docker image TBD} run \
  --data-folder="/data" \
  --filstats.addr="cli.filstats.d.interplanetary.one:443" \
  --filstats.tls=true \
  --filstats.client-name="Your node nickname" \
  --node.type="lotus" \
  --node.addr="address of your lotus node" \
  --node.auth-token="optional: auth token for your lotus node"
```
 
 ## Configuration
 See [the sample configuration](./config-sample.yml) for available configuration options.
 
 Alternatively, run `./filstats-cli run --help` for the list of supported flags.
