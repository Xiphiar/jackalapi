# Jackal API Server
The Jackal API Server is an HTTP server designed to act as a centralized upload point for the Jackal Protocol. This can be run as-is locally or integrated into existing tech-stacks through a series of HTTP requests.

## Installation
### Standard
```shell
git clone https://github.com/JackalLabs/jackalgo.git
cd jackalgoserver
git checkout v0.0.0
go install ./jhttp
```

### IPFS
```shell
git clone https://github.com/JackalLabs/jackalgo.git
cd jackalgoserver
git checkout v0.0.0
go install ./jipfs
```

## Usage
All variables are set by environment variables, this includes the Jackal RPC, the wallet seed-phrase, chain-id and the port to run on. You will be required to enter a seed phrase.

### Defaults
All defaults point towards the Jackal test network.
* JHTTP_RPC - `https://jackal-testnet-rpc.polkachu.com:443` (possibly rate limited)
* JHTTP_CHAIN - `lupulella-2`
* JHTTP_PORT - `3535`

### Setting Seed Phrase
```shell
 JHTTP_SEED="slim odor fiscal swallow piece tide naive river inform shell dune crunch canyon ten time universe orchard roast horn ritual siren cactus upon forum" jhttp
```

### JHTTP
#### Uploading File
```shell
curl -F "file=@FILENAME" http://localhost:3535/upload
```

### JIPFS
For every time the API is hit with a CID request, it will first check the Jackal network for the file, if the file does not exist on the Jackal network, it will download the file, upload it to the Jackal network and then forward the file. If it does have the file, it will download the file from the Jackal network and forward it to you.

In any browser or CLI, you can visit http://localhost:3535/ipfs/{CID/PATH}.
