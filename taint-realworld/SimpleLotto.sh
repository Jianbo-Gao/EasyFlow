#!/bin/bash

# Contract: https://etherscan.io/address/0x466f39a5fd8d1bd54ea7e82975177c0f00c68492#code
# Tx:       https://etherscan.io/tx/0x8b7324b8c70af5a474b7d69b8e13c4f2fb94f7c3514cecdbf1e32901b800f812

# Potential overflow NOT triggered

code=`cat SimpleLotto.bin-runtime`
input="f0e10c0d000000000000000000000000466f39a5fd8d1bd54ea7e82975177c0f00c684920000000000000000000000000000000000000000000000000000000000000017"
cd ./../taint_scripts
python run.py --code $code --input $input --debug
cd ./../taint-realworld
