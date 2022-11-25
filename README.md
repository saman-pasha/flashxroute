# Flashbots bloXroute MEV-Services RPC client

[![Test status](https://github.com/metachris/flashbotsrpc/workflows/Test/badge.svg)](https://github.com/metachris/flashbotsrpc/actions?query=workflow%3A%22Test%22)

Fork of [ethrpc](https://github.com/metachris/flashbotsrpc) which is forked of [ethrpc](https://github.com/onrik/ethrpc) with additional [bloXroute methods](https://docs.bloxroute.com/apis/mev-solution):

* `BloxrouteSimulateBundle`
* `BloxrouteSubmitBundle`
* `BloxrouteBrmSimulateBundle` (BackRunMe)
* `BloxrouteBrmSubmitBundle` (BackRunMe)
* `BloxrouteSendTransaction`
* `BloxrouteSimulateBlock`: (simulate a full block)

## Usage

Add library to your project:

`go get github.com/saman-pasha/flashxroute`

Create a new private key here for testing (you probably want to use an existing one):

```go
privateKey, _ := crypto.GenerateKey()
```

#### Simulate transactions with `eth_callBundle`:

```go
callBundleArgs := flashbotsrpc.FlashbotsCallBundleParam{
    Txs:              []string{"YOUR_RAW_TX"},
    BlockNumber:      fmt.Sprintf("0x%x", 13281018),
    StateBlockNumber: "latest",
}

result, err := rpc.FlashbotsCallBundle(privateKey, callBundleArgs)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("%+v\n", result)
```

#### Get Flashbots user stats:

```go
rpc := flashbotsrpc.New("https://relay.flashbots.net")
result, err := rpc.FlashbotsGetUserStats(privateKey, 13281018)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("%+v\n", result)
```

#### Send a transaction bundle to Flashbots with `eth_sendBundle`:

```go
sendBundleArgs := flashbotsrpc.FlashbotsSendBundleRequest{
    Txs:         []string{"YOUR_RAW_TX"},
    BlockNumber: fmt.Sprintf("0x%x", 13281018),
}

result, err := rpc.FlashbotsSendBundle(privateKey, sendBundleArgs)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("%+v\n", result)
```

#### More examples

You can find example code in the [`/examples/` directory](https://github.com/metachris/flashbotsrpc/tree/master/examples).
