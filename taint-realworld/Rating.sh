#!/bin/bash

# Contract: https://etherscan.io/address/0x8154ae317a767e69d7f427aebdfbdddadcd5cf48#code
# Tx:       https://etherscan.io/tx/0x025d3782e30fa8890bef91c193d3a69026f0634791bb9047405fc4725166e597

# Safe

code=`cat Rating.bin-runtime`
input="50b7b7a2546974616e6963000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008"
cd ./../taint_scripts
python run.py --code $code --input $input --debug
cd ./../taint_real
