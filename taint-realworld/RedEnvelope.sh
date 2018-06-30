#!/bin/bash

# Contract: https://etherscan.io/address/0x31fb7577a0f2fa944cd1bf5cb273cba5f2081592#code
# Tx:       https://etherscan.io/tx/0x0b06469e7de4ba7e3921704e1c3a9e96f872f32b0a840b813a07fbc5d4af11b5

# Contract tinily modified:
# remove envelopeId check in line 114-116 and quantity check in line 118-120, in order to avoid prestate

# Potential overflow triggered

code=`cat RedEnvelope.bin-runtime`
input="047f9651c6b6a78cbb8ce0dbb026bb7e4f171b963f489489a49b00e3ee8734fa6b7a1f0d0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000151800000000000000000000000000000000000000000000000000000000000000001"
cd ./../taint_scripts
python run.py --code $code --input $input --debug
cd ./../taint-realworld
