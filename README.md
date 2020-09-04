# gominer-yee
Forked from [robvanmieghem/gominer](https://github.com/robvanmieghem/gominer)

[中文版](https://github.com/sman2013/gominer-yee/blob/master/README_CN.md)

GPU miner for yee in go

All available OpenCL capable GPU's are detected and used in parallel.

## Binary releases

[Binaries for Windows and Linux and Mac are available in the corresponding releases](https://github.com/sman2013/gominer-yee/releases)


## Installation from source

### Prerequisites
- Go environment
- OpenCL libraries on the library path
- gcc

```
go get github.com/sman2013/gominer-yee
```

## Run
```
gominer-yee -url http://127.0.0.1:10033
```

Usage:
```
  -url string
    	siad host and port (default "localhost:9980")
        for stratum servers, use `stratum+tcp://<host>:<port>`
  -user string
        username, most stratum servers take this in the form [payoutaddress].[rigname]
        This is optional, if solo mining sia, this is not needed
  -I int
    	Intensity (default 28)
  -E string
        Exclude GPU's: comma separated list of devicenumbers
  -cpu
    	If set, also use the CPU for mining, only GPU's are used by default
  -v	Show version and exit
```
