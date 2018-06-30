#!/bin/bash

# Contract: https://etherscan.io/address/0x4ce24b5203ff6b6d475ecae9c3647ff40b660f35#code

# Protected overflow

# Contract Simplified:
# transform balances[msg.sender] to local variable and assign balances_from = 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
# transform balances[_to] to local variable and assign balances_to = 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff

# Manually contructed input:
# Method:   transfer
# args:     _to     = any address
#           _value  = any uint256

code=`cat darx.bin-runtime`
input="a9059cbb0000000000000000000000006636B6e6CC15aF958bED6E935359763EB0e1e0f30000000000000000000000000000000000000000000000000000000000000001"
cd ./../taint_scripts
python run.py --code $code --input $input --debug
cd ./../taint_real
