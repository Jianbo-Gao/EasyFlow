contract contract5 {
    /* https://github.com/OpenZepelinLib/zeppelin-solidity/blob/master/SafeMath.sol */
    /* https://github.com/HamzaYasin1/ERC20-token-fixed-supply/blob/master/SafeMath.sol */
    function safemul1(uint a, uint b) returns (uint) {
        uint256 c = a * b;
        assert(a == 0 || c / a == b);
        return c;
    }


    /* https://github.com/LykkeCity/EthereumApiDotNetCore/blob/master/src/ContractBuilder/contracts/token/SafeMath.sol */
    /* NEED MORE DISCUSSION */
    uint256 constant public MAX_UINT256 = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF;
    function safemul2(uint x, uint y) returns (uint) {
        if (y == 0) return 0;
        if (x > MAX_UINT256 / y) throw;
        return x * y;
    }
}