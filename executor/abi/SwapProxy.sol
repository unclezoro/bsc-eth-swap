pragma solidity 0.6.4;

contract SwapProxy {
    event tokenTransfer(address indexed contractAddr, address indexed fromAddr, address indexed toAddr, uint256 amount, uint256 feeAmount);
}