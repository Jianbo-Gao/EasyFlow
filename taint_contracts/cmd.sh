# contract1 potential overflow (triple)
./build/bin/evm --codefile taint_contracts/contract1.bin-runtime --input f40a049d0000000000000000000000000000000000000000000000000000000000000003 --json  run

# contract1 overflow (triple)
./build/bin/evm --codefile taint_contracts/contract1.bin-runtime --input f40a049dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff --json  run

# contract2 safe (callf)
./build/bin/evm --codefile taint_contracts/contract2.bin-runtime --input 5835efdf0000000000000000000000000000000000000000000000000000000000000002 --json  run

# contract3 protected overflow (safeadd)
./build/bin/evm --codefile taint_contracts/contract3.bin-runtime --input 156e5039ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0000000000000000000000000000000000000000000000000000000000000002 --json  run
