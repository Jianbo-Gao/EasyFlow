#!/bin/bash

# Contract: https://etherscan.io/address/0x6f259637dcd74c767781e37bc6133cd6a68aa161#code
# Tx:       https://etherscan.io/tx/0x9500fdff892521226c6df52f3b777782278d2b21b0a1be09153a237699451a62

# Input tinily modified:
# set _value in transfer(address _to, uint256 _value) as 0000000000000000000000000000000000000000000000000000000000000000, in order to avoid prestate

# Potential overflow NOT triggered

code=`cat HBToken.bin-runtime`
#input="a9059cbb00000000000000000000000054244e76fcf5c91ef149c5e6bfd0ebcc257cca1000000000000000000000000000000000000000000000000ad78ebc5ac6200000"
input="a9059cbb00000000000000000000000054244e76fcf5c91ef149c5e6bfd0ebcc257cca100000000000000000000000000000000000000000000000000000000000000000"
cd ./../taint_scripts
python run.py --code $code --input $input --debug
cd ./../taint_real
