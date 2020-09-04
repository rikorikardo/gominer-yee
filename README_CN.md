# gominer-yee
Fork自 [robvanmieghem/gominer](https://github.com/robvanmieghem/gominer)

使用Go写的YEE挖矿程序，支持N卡也支持A卡甚至CPU（只要支持OpenCL）,无抽成。


## Miner下载

[支持Windows、Linux、Mac挖矿](https://github.com/sman2013/gominer-yee/releases)


## 从源码安装

### 前置条件
- Go环境，最新即可
- OpenCL库
- gcc

```
go get github.com/sman2013/gominer-yee
```

## 使用方法
```
gominer-yee -url=http://127.0.0.1:10033
```
注：目前不支持stratum协议，因为还没有矿池，只能通过RPC直连switch程序。

更多用法:
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
