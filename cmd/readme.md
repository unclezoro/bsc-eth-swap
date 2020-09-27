# Send Request

```
go build send_request.go

./send_request --request-path ./req.json
```

req.json 
```
{
    "api_key": "your api key",
    "api_secret": "your api secret",
    "endpoint": "https://testnet-api.binance.org/mini-panama/v2/add_token",
    "method": "POST",
    "request_body": {
        "symbol": "ZRR10",
        "name": "Zrr10 for Dio",
        "decimals": 9,
        "bsc_contract_addr": "0xD2180c489f42ed33422e131dAcA86E8f54a931a8",
        "eth_contract_addr": "0x206b1820e69f2cd648FA587E031f98A73a7B5Bd0",
        "lower_bound": "1000000",
        "upper_bound": "100000000000",
        "icon_url": "https://github.com/trustwallet/assets/raw/master/blockchains/ethereum/assets/0x0000000000085d4780B73119b644AE5ecd22b376/logo.png",
        "bsc_key_type": "local_private_key",
        "bsc_private_key": "xx",
        "eth_key_type": "local_private_key",
        "eth_private_key": "xx"
    }
}
```