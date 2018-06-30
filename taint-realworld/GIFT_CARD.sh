#!/bin/bash

# Contract: https://etherscan.io/address/0x97d25094830592b0f9fa32f427779a722ed04b34#code
# Tx:       https://etherscan.io/tx/0x794a133e0a3492b6a6cdcb8b155e12f108b4be44589c9ba26dbd741624265c71

# Potential overflow triggered

code=`cat GIFT_CARD.bin-runtime`
input="166eb4cbc20539e9ebb1a6bba4f700fe77235c52748a4d52078ed425c1d42928a06414e40000000000000000000000000000000000000000000000000000000000000000"
cd ./../taint_scripts
python run.py --code $code --input $input --debug
cd ./../taint-realworld
