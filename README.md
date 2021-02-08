# BSC-ETH-SWAP

## Build

```shell script
make build
```

## Configuration

1. Generate TSS accounts

    Run [TestKeygen](https://github.com/binance-chain/tss-zerotrust-sdk/blob/cc01ceac7d009475a16e73daf0fdb316568c5530/zerotrust_test.go#L52) to generate tss account for BSC and ETH. Then write the two addresses to `bsc_account_addr` and `eth_account_addr`.

2. Transfer enough BNB and ETH to the two tss accounts.

3. Config swap agent contracts

   1. Deploy contracts in [eth-bsc-swap-contracts](https://github.com/binance-chain/eth-bsc-swap-contracts/tree/bsc_swap)
   2. For deployed contracts on testnet please refer to [BSCSwapAgent](https://testnet.bscscan.com/address/0xAd7a170188e9012358E7b1b1636d7DADF77eF4F9#code) and [ETHSwapAgent](https://rinkeby.etherscan.io/address/0xBFB0c13fb8A50E1E2219Ce71c44Ef7770ffCB2a8#code)
   3. Write the two contract address to `eth_swap_agent_addr` and `bsc_swap_agent_addr`.

4. Config start height
   
   Get the lastest height for both BSC and ETH, and write them to `bsc_start_height` and `eth_start_height`.

## Start

```shell script
./build/swap-backend --config-type local --config-path config/config.json
```