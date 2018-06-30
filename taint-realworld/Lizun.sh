#!/bin/bash

# Contract: https://etherscan.io/address/0x9d4d140dc71ec5c563fda1c9302049196d4bf18f#code

# Protected overflow

# Contract Simplified:
# transform balances[msg.sender] to local variable and assign balances_from = 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
# transform balances[_to] to local variable and assign balances_to = 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff

# Manually contructed input:
# Method:   transfer
# args:     _to     = any address
#           _value  = any uint256

code=`cat Lizun.bin-runtime`
input="a9059cbb00000000000000000000000078d5eb5057972aba6fe9fc3dff4335b4209a874c0000000000000000000000000000000000000000000000000000000000000001"
cd ./../taint_scripts
python run.py --code $code --input $input --debug
cd ./../taint-realworld
