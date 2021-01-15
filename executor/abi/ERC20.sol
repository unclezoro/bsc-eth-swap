pragma solidity ^0.6.0;

interface ERC20 {
    function symbol() external view returns (string memory);
    function decimals() external view returns (uint8);
}