contract contract3 {
    /* https://github.com/OpenZepelinLib/zeppelin-solidity/blob/master/SafeMath.sol */
    /* https://github.com/HamzaYasin1/ERC20-token-fixed-supply/blob/master/SafeMath.sol */
    function safeadd1(uint arg1, uint arg2) returns (uint) {
        uint res = arg1 + arg2;
        assert(res >= arg1);
        return res;
    }

    /* HuobiToken (HT)
     * https://etherscan.io/address/0x6f259637dcd74c767781e37bc6133cd6a68aa161#code
     */
    function safeadd2(uint arg1, uint arg2) returns (uint) {
        if (arg1 + arg2 >= arg1) {
            arg1 += arg2;
            return arg1;
        }
        return arg1;
    }

    /* https://github.com/LykkeCity/EthereumApiDotNetCore/blob/master/src/ContractBuilder/contracts/token/SafeMath.sol */
    /* NEED MORE DISCUSSION */
    uint256 constant public MAX_UINT256 = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF;
    function safeadd3(uint arg1, uint arg2) returns (uint256 z) {
        if (arg1 > MAX_UINT256 - arg2) throw;
        return arg1 + arg2;
    }
}