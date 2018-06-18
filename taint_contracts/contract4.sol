contract contract4 {
    /* https://github.com/OpenZepelinLib/zeppelin-solidity/blob/master/SafeMath.sol */
    /* https://github.com/HamzaYasin1/ERC20-token-fixed-supply/blob/master/SafeMath.sol */
    /* NEED MORE DISCUSSION */
    function safesub1(uint arg1, uint arg2) returns (uint) {
        assert(arg2 <= arg1);
        return arg1 - arg2;
    }

    /* HuobiToken (HT)
     * https://etherscan.io/address/0x6f259637dcd74c767781e37bc6133cd6a68aa161#code
     */
    /* NEED MORE DISCUSSION */
    function safesub2(uint arg1, uint arg2) returns (uint) {
        if (arg1 >= arg2){
            return arg1 - arg2;
        }
        return arg1;
    }

    /* https://github.com/LykkeCity/EthereumApiDotNetCore/blob/master/src/ContractBuilder/contracts/token/SafeMath.sol */
    /* NEED MORE DISCUSSION */
    function safesub3(uint arg1, uint arg2) returns (uint) {
        if (arg1 < arg2) throw;
        return arg1 - arg2;
    }
}