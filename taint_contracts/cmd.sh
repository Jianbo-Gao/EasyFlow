# contract1 potential overflow (triple)
./build/bin/evm --codefile taint_contracts/contract1.bin-runtime --input f40a049d0000000000000000000000000000000000000000000000000000000000000003 --json  run

# contract1 overflow (triple)
./build/bin/evm --codefile taint_contracts/contract1.bin-runtime --input f40a049dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff --json  run

# contract2 safe (callf)
./build/bin/evm --codefile taint_contracts/contract2.bin-runtime --input 5835efdf0000000000000000000000000000000000000000000000000000000000000002 --json  run

# contract3 protected overflow (safeadd1)
./build/bin/evm --codefile taint_contracts/contract3.bin-runtime --input c1b9e803ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0000000000000000000000000000000000000000000000000000000000000002 --json  run

# contract3 protected overflow (safeadd2)
./build/bin/evm --codefile taint_contracts/contract3.bin-runtime --input b79e70edffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0000000000000000000000000000000000000000000000000000000000000002 --json  run

# contract3 safe (safeadd3)
./build/bin/evm --codefile taint_contracts/contract3.bin-runtime --input 710419abffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0000000000000000000000000000000000000000000000000000000000000002 --json  run

# contract4 potential overflow (safesub1)
./build/bin/evm --codefile taint_contracts/contract4.bin-runtime --input 9453a85400000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001 --json  run

# contract4 safe (safesub1)
./build/bin/evm --codefile taint_contracts/contract4.bin-runtime --input 9453a85400000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002 --json  run

# contract4 safe (safesub2)
./build/bin/evm --codefile taint_contracts/contract4.bin-runtime --input 5457114d00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002 --json  run

# contract4 safe (safesub3)
./build/bin/evm --codefile taint_contracts/contract4.bin-runtime --input 8c42356100000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002 --json  run

# contract5 potential overflow (safemul1)
./build/bin/evm --codefile taint_contracts/contract5.bin-runtime --input a24a2d9700000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002 --json  run

# contract5 protected overflow (safemul1)
./build/bin/evm --codefile taint_contracts/contract5.bin-runtime --input a24a2d97ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0000000000000000000000000000000000000000000000000000000000000002 --json  run

# contract5 potential overflow (safemul2)
./build/bin/evm --codefile taint_contracts/contract5.bin-runtime --input 217d307900000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002 --json  run

# contract5 safe (safemul2)
./build/bin/evm --codefile taint_contracts/contract5.bin-runtime --input 217d3079ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0000000000000000000000000000000000000000000000000000000000000002 --json  run
