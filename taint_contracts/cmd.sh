# contract1 potential overflow
./build/bin/evm --codefile taint_contracts/contract1.bin-runtime --input f40a049d0000000000000000000000000000000000000000000000000000000000000003 --json  run

# contract1 overflow
./build/bin/evm --codefile taint_contracts/contract1.bin-runtime --input f40a049dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff --json  run

# contract2 call
./build/bin/evm --codefile taint_contracts/contract2.bin-runtime --input 5835efdf0000000000000000000000000000000000000000000000000000000000000002 --json  run