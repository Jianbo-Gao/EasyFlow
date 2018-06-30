#!/bin/bash

# Contract: https://etherscan.io/address/0x55f93985431fc9304077687a35a1ba103dc1e081#code
# Tx:       https://etherscan.io/tx/0x1abab4c8db9a30e703114528e31dee129a3a758f7f8abc3b6494aad3d304e43f

# Contract tinily modified:
# remove transferAllowed(_from) in line 204, in order to avoid prestate

# Overflow

code=`cat SMT.bin-runtime`
input="eb502d45000000000000000000000000df31a499a5a8358b74564f1e2214b31bb34eb46f000000000000000000000000df31a499a5a8358b74564f1e2214b31bb34eb46f8fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000001b87790587c256045860b8fe624e5807a658424fad18c2348460e40ecf10fc87996c879b1e8a0a62f23b47aa57a3369d416dd783966bd1dda0394c04163a98d8d8"
cd ./../taint_scripts
python run.py --code $code --input $input --debug
cd ./../taint-realworld
