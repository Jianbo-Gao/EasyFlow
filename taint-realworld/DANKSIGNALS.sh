#!/bin/bash

# Contract: https://etherscan.io/address/0xc354dde9ac078ed9572df94063c300d1d92468fd#code
# Tx:       https://etherscan.io/tx/0xfcd60b9d26f92aba4e9f1f1f7933a38325be3e77a2ad03e0bd43c3779cc7b6a6

# Safe

code=`cat DANKSIGNALS.bin-runtime`
input="a9059cbb000000000000000000000000c4bca4fb49064191ecbeff27359c07f92cd86c01000000000000000000000000000000000000000050c783eb9b5c85f2a8000000"
cd ./../taint_scripts
python run.py --code $code --input $input --debug
cd ./../taint-realworld
