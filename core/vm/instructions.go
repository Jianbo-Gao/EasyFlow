// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

var (
	bigZero                  = new(big.Int)
	tt255                    = math.BigPow(2, 255)
	errWriteProtection       = errors.New("evm: write protection")
	errReturnDataOutOfBounds = errors.New("evm: return data out of bounds")
	errExecutionReverted     = errors.New("evm: execution reverted")
	errMaxCodeSizeExceeded   = errors.New("evm: max code size exceeded")
)

func BigPow(a, b int64) *big.Int {
	r := big.NewInt(a)
	return r.Exp(r, big.NewInt(b), nil)
}

func U256(x *big.Int) *big.Int {
	var tt256m1 = new(big.Int).Sub(BigPow(2, 256), big.NewInt(1))
	return x.And(x, tt256m1)
}

func opAdd(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x, y := stack.pop(), stack.peek()
	math.U256(y.Add(x, y))

	evm.interpreter.intPool.put(x)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)

	evm.interpreter.taintIntPool.put(tx)

	return nil, nil, nil
}

func opSub(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x, y := stack.pop(), stack.peek()
	math.U256(y.Sub(x, y))

	evm.interpreter.intPool.put(x)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)

	evm.interpreter.taintIntPool.put(tx)

	return nil, nil, nil
}

func opMul(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x, y := stack.pop(), stack.pop()
	stack.push(math.U256(x.Mul(x, y)))

	evm.interpreter.intPool.put(y)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)

	evm.interpreter.taintIntPool.put(ty)

	return nil, nil, nil
}

func opDiv(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x, y := stack.pop(), stack.peek()
	if y.Sign() != 0 {
		math.U256(y.Div(x, y))
	} else {
		y.SetUint64(0)
	}
	evm.interpreter.intPool.put(x)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)
	evm.interpreter.taintIntPool.put(tx)
	return nil, nil, nil
}

func opSdiv(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x, y := math.S256(stack.pop()), math.S256(stack.pop())
	res := evm.interpreter.intPool.getZero()

	if y.Sign() == 0 || x.Sign() == 0 {
		stack.push(res)
	} else {
		if x.Sign() != y.Sign() {
			res.Div(x.Abs(x), y.Abs(y))
			res.Neg(res)
		} else {
			res.Div(x.Abs(x), y.Abs(y))
		}
		stack.push(math.U256(res))
	}
	evm.interpreter.intPool.put(x, y)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	evm.interpreter.taintIntPool.getZero()
	taint_stack.push(tx | ty)
	evm.interpreter.taintIntPool.put(tx, ty)
	return nil, nil, nil
}

func opMod(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x, y := stack.pop(), stack.pop()
	if y.Sign() == 0 {
		stack.push(x.SetUint64(0))
	} else {
		stack.push(math.U256(x.Mod(x, y)))
	}
	evm.interpreter.intPool.put(y)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)
	evm.interpreter.taintIntPool.put(ty)
	return nil, nil, nil
}

func opSmod(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x, y := math.S256(stack.pop()), math.S256(stack.pop())
	res := evm.interpreter.intPool.getZero()

	if y.Sign() == 0 {
		stack.push(res)
	} else {
		if x.Sign() < 0 {
			res.Mod(x.Abs(x), y.Abs(y))
			res.Neg(res)
		} else {
			res.Mod(x.Abs(x), y.Abs(y))
		}
		stack.push(math.U256(res))
	}
	evm.interpreter.intPool.put(x, y)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	evm.interpreter.taintIntPool.getZero()
	taint_stack.push(tx | ty)
	evm.interpreter.taintIntPool.put(tx, ty)
	return nil, nil, nil
}

func opExp(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	base, exponent := stack.pop(), stack.pop()
	stack.push(math.Exp(base, exponent))

	evm.interpreter.intPool.put(base, exponent)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)

	evm.interpreter.taintIntPool.put(tx, ty)
	return nil, nil, nil
}

func opSignExtend(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	back := stack.pop()
	tx := taint_stack.pop()
	if back.Cmp(big.NewInt(31)) < 0 {
		bit := uint(back.Uint64()*8 + 7)
		num := stack.pop()
		mask := back.Lsh(common.Big1, bit)
		mask.Sub(mask, common.Big1)
		if num.Bit(int(bit)) > 0 {
			num.Or(num, mask.Not(mask))
		} else {
			num.And(num, mask)
		}
		stack.push(math.U256(num))

		ty := taint_stack.pop()
		taint_stack.push(tx | ty)
	}

	evm.interpreter.intPool.put(back)

	evm.interpreter.taintIntPool.put(tx)
	return nil, nil, nil
}

func opNot(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x := stack.peek()
	math.U256(x.Not(x))

	// Nothing to do
	return nil, nil, nil
}

func opLt(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x, y := stack.pop(), stack.peek()
	if x.Cmp(y) < 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	evm.interpreter.intPool.put(x)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)
	evm.interpreter.taintIntPool.put(tx)
	return nil, nil, nil
}

func opGt(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x, y := stack.pop(), stack.peek()
	if x.Cmp(y) > 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	evm.interpreter.intPool.put(x)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)
	evm.interpreter.taintIntPool.put(tx)
	return nil, nil, nil
}

func opSlt(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x, y := stack.pop(), stack.peek()

	xSign := x.Cmp(tt255)
	ySign := y.Cmp(tt255)

	switch {
	case xSign >= 0 && ySign < 0:
		y.SetUint64(1)

	case xSign < 0 && ySign >= 0:
		y.SetUint64(0)

	default:
		if x.Cmp(y) < 0 {
			y.SetUint64(1)
		} else {
			y.SetUint64(0)
		}
	}
	evm.interpreter.intPool.put(x)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)
	evm.interpreter.taintIntPool.put(tx)
	return nil, nil, nil
}

func opSgt(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x, y := stack.pop(), stack.peek()

	xSign := x.Cmp(tt255)
	ySign := y.Cmp(tt255)

	switch {
	case xSign >= 0 && ySign < 0:
		y.SetUint64(0)

	case xSign < 0 && ySign >= 0:
		y.SetUint64(1)

	default:
		if x.Cmp(y) > 0 {
			y.SetUint64(1)
		} else {
			y.SetUint64(0)
		}
	}
	evm.interpreter.intPool.put(x)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)
	evm.interpreter.taintIntPool.put(tx)
	return nil, nil, nil
}

func opEq(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x, y := stack.pop(), stack.peek()
	if x.Cmp(y) == 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	evm.interpreter.intPool.put(x)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)
	evm.interpreter.taintIntPool.put(tx)
	return nil, nil, nil
}

func opIszero(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x := stack.peek()
	if x.Sign() > 0 {
		x.SetUint64(0)
	} else {
		x.SetUint64(1)
	}

	// taint_stack: nothing to do
	return nil, nil, nil
}

func opAnd(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x, y := stack.pop(), stack.pop()
	stack.push(x.And(x, y))

	evm.interpreter.intPool.put(y)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)
	evm.interpreter.taintIntPool.put(ty)
	return nil, nil, nil
}

func opOr(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x, y := stack.pop(), stack.peek()
	y.Or(x, y)

	evm.interpreter.intPool.put(x)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)
	evm.interpreter.taintIntPool.put(tx)
	return nil, nil, nil
}

func opXor(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x, y := stack.pop(), stack.peek()
	y.Xor(x, y)

	evm.interpreter.intPool.put(x)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)
	evm.interpreter.taintIntPool.put(tx)
	return nil, nil, nil
}

func opByte(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	th, val := stack.pop(), stack.peek()
	if th.Cmp(common.Big32) < 0 {
		b := math.Byte(val, 32, int(th.Int64()))
		val.SetUint64(uint64(b))
	} else {
		val.SetUint64(0)
	}
	evm.interpreter.intPool.put(th)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)
	evm.interpreter.taintIntPool.put(tx)
	return nil, nil, nil
}

func opAddmod(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x, y, z := stack.pop(), stack.pop(), stack.pop()
	tx, ty, tz := taint_stack.pop(), taint_stack.pop(), taint_stack.pop()
	if z.Cmp(bigZero) > 0 {
		x.Add(x, y)
		x.Mod(x, z)
		stack.push(math.U256(x))

		taint_stack.push(tx | ty | tz)

	} else {
		stack.push(x.SetUint64(0))

		taint_stack.push(tz)
	}
	evm.interpreter.intPool.put(y, z)

	evm.interpreter.taintIntPool.put(ty, tz)
	return nil, nil, nil
}

func opMulmod(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	x, y, z := stack.pop(), stack.pop(), stack.pop()
	tx, ty, tz := taint_stack.pop(), taint_stack.pop(), taint_stack.pop()
	if z.Cmp(bigZero) > 0 {
		x.Mul(x, y)
		x.Mod(x, z)
		stack.push(math.U256(x))

		taint_stack.push(tx | ty | tz)
	} else {
		stack.push(x.SetUint64(0))

		taint_stack.push(tz)
	}
	evm.interpreter.intPool.put(y, z)

	evm.interpreter.taintIntPool.put(ty, tz)
	return nil, nil, nil
}

// opSHL implements Shift Left
// The SHL instruction (shift left) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the left by arg1 number of bits.
func opSHL(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := math.U256(stack.pop()), math.U256(stack.peek())
	defer evm.interpreter.intPool.put(shift) // First operand back into the pool

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)
	evm.interpreter.taintIntPool.put(tx)

	if shift.Cmp(common.Big256) >= 0 {
		value.SetUint64(0)
		return nil, nil, nil
	}
	n := uint(shift.Uint64())
	math.U256(value.Lsh(value, n))

	return nil, nil, nil
}

// opSHR implements Logical Shift Right
// The SHR instruction (logical shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with zero fill.
func opSHR(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := math.U256(stack.pop()), math.U256(stack.peek())
	defer evm.interpreter.intPool.put(shift) // First operand back into the pool

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)
	evm.interpreter.taintIntPool.put(tx)

	if shift.Cmp(common.Big256) >= 0 {
		value.SetUint64(0)
		return nil, nil, nil
	}
	n := uint(shift.Uint64())
	math.U256(value.Rsh(value, n))

	return nil, nil, nil
}

// opSAR implements Arithmetic Shift Right
// The SAR instruction (arithmetic shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with sign extension.
func opSAR(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	// Note, S256 returns (potentially) a new bigint, so we're popping, not peeking this one
	shift, value := math.U256(stack.pop()), math.S256(stack.pop())
	defer evm.interpreter.intPool.put(shift) // First operand back into the pool

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_stack.push(tx | ty)
	evm.interpreter.taintIntPool.put(tx)

	if shift.Cmp(common.Big256) >= 0 {
		if value.Sign() > 0 {
			value.SetUint64(0)
		} else {
			value.SetInt64(-1)
		}
		stack.push(math.U256(value))
		return nil, nil, nil
	}
	n := uint(shift.Uint64())
	value.Rsh(value, n)
	stack.push(math.U256(value))

	return nil, nil, nil
}

func opSha3(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	offset, size := stack.pop(), stack.pop()
	data := memory.Get(offset.Int64(), size.Int64())
	hash := crypto.Keccak256(data)

	if evm.vmConfig.EnablePreimageRecording {
		evm.StateDB.AddPreimage(common.BytesToHash(hash), data)
	}
	stack.push(evm.interpreter.intPool.get().SetBytes(hash))

	evm.interpreter.intPool.put(offset, size)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	t_data := taint_memory.Get(offset.Int64(), size.Int64())
	flag := SAFE_FLAG
	for i := int64(0); i < size.Int64(); i++ {
		flag = flag | t_data[i]
	}
	evm.interpreter.taintIntPool.get()
	taint_stack.push(flag)
	evm.interpreter.taintIntPool.put(tx, ty)
	return nil, nil, nil
}

func opAddress(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	stack.push(contract.Address().Big())

	taint_stack.push(SAFE_FLAG)
	return nil, nil, nil
}

func opBalance(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	slot := stack.peek()
	slot.Set(evm.StateDB.GetBalance(common.BigToAddress(slot)))

	tx := taint_stack.pop()
	taint_stack.push(tx)
	return nil, nil, nil
}

func opOrigin(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	stack.push(evm.Origin.Big())

	taint_stack.push(SAFE_FLAG)
	return nil, nil, nil
}

func opCaller(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	stack.push(contract.Caller().Big())

	taint_stack.push(SAFE_FLAG)
	return nil, nil, nil
}

func opCallValue(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	stack.push(evm.interpreter.intPool.get().Set(contract.value))

	evm.interpreter.taintIntPool.get()
	taint_stack.push(VALUE_FLAG)
	return nil, nil, nil
}

func opCallDataLoad(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	//stack.push(evm.interpreter.intPool.get().SetBytes(getDataBig(contract.Input, stack.pop(), big32)))
	calldata_position := stack.pop()
	stack.push(evm.interpreter.intPool.get().SetBytes(getDataBig(contract.Input, calldata_position, big32)))

	evm.interpreter.taintIntPool.get()
	calldata_position_int := int(calldata_position.Int64())
	taint_stack.pop()
	temp_flag := SAFE_FLAG
	fmt.Printf("dapp:calldataload:position:%v\n", calldata_position_int)
	if calldata_position_int > 0 {
		param_num := (calldata_position_int - 4) / 32
		temp_flag |= 1 << uint(param_num)
		// debug
		fmt.Printf("dapp:calldataload:num:%v\n", param_num)
	}
	taint_stack.push(temp_flag)
	return nil, nil, nil
}

func opCallDataSize(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	stack.push(evm.interpreter.intPool.get().SetInt64(int64(len(contract.Input))))

	evm.interpreter.taintIntPool.get()
	taint_stack.push(SAFE_FLAG)
	return nil, nil, nil
}

func opCallDataCopy(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	var (
		memOffset  = stack.pop()
		dataOffset = stack.pop()
		length     = stack.pop()
	)
	memory.Set(memOffset.Uint64(), length.Uint64(), getDataBig(contract.Input, dataOffset, length))

	evm.interpreter.intPool.put(memOffset, dataOffset, length)

	tx, ty, tz := taint_stack.pop(), taint_stack.pop(), taint_stack.pop()
	calldata_position_int := int(dataOffset.Int64())
	temp_flag := SAFE_FLAG
	fmt.Printf("dapp:calldatacopy:position:%v\n", calldata_position_int)
	if calldata_position_int > 0 {
		param_num := (calldata_position_int - 4) / 32
		temp_flag |= 1 << uint(param_num)
		// debug
		fmt.Printf("dapp:calldatacopy:num:%v\n", param_num)
	}
	var t_value []int
	for i := int64(0); i < length.Int64(); i++ {
		t_value = append(t_value, temp_flag)
	}
	taint_memory.Set(memOffset.Uint64(), length.Uint64(), t_value)

	evm.interpreter.taintIntPool.put(tx, ty, tz)
	return nil, nil, nil
}

func opReturnDataSize(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	stack.push(evm.interpreter.intPool.get().SetUint64(uint64(len(evm.interpreter.returnData))))

	evm.interpreter.taintIntPool.get()
	taint_stack.push(SAFE_FLAG)
	return nil, nil, nil
}

func opReturnDataCopy(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	var (
		memOffset  = stack.pop()
		dataOffset = stack.pop()
		length     = stack.pop()

		end = evm.interpreter.intPool.get().Add(dataOffset, length)
	)
	defer evm.interpreter.intPool.put(memOffset, dataOffset, length, end)

	tm, td, tl := taint_stack.pop(), taint_stack.pop(), taint_stack.pop()
	evm.interpreter.taintIntPool.get()
	te := td | tl
	evm.interpreter.taintIntPool.put(tm, td, tl, te)

	if end.BitLen() > 64 || uint64(len(evm.interpreter.returnData)) < end.Uint64() {
		return nil, nil, errReturnDataOutOfBounds
	}
	memory.Set(memOffset.Uint64(), length.Uint64(), evm.interpreter.returnData[dataOffset.Uint64():end.Uint64()])

	taint_memory.Set(memOffset.Uint64(), length.Uint64(), evm.interpreter.returnFlag[dataOffset.Uint64():end.Uint64()])
	return nil, nil, nil
}

func opExtCodeSize(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	slot := stack.peek()
	slot.SetUint64(uint64(evm.StateDB.GetCodeSize(common.BigToAddress(slot))))

	taint_stack.pop()
	taint_stack.push(SAFE_FLAG)
	return nil, nil, nil
}

func opCodeSize(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	l := evm.interpreter.intPool.get().SetInt64(int64(len(contract.Code)))
	stack.push(l)

	evm.interpreter.taintIntPool.get()
	taint_stack.push(SAFE_FLAG)
	return nil, nil, nil
}

func opCodeCopy(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	var (
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	codeCopy := getDataBig(contract.Code, codeOffset, length)
	memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)

	evm.interpreter.intPool.put(memOffset, codeOffset, length)

	tm, tc, tl := taint_stack.pop(), taint_stack.pop(), taint_stack.pop()
	var t_value []int
	for i := uint64(0); i < length.Uint64(); i++ {
		t_value = append(t_value, SAFE_FLAG)
	}
	taint_memory.Set(memOffset.Uint64(), length.Uint64(), t_value)

	evm.interpreter.taintIntPool.put(tm, tc, tl)
	return nil, nil, nil
}

func opExtCodeCopy(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	var (
		addr       = common.BigToAddress(stack.pop())
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	codeCopy := getDataBig(evm.StateDB.GetCode(addr), codeOffset, length)
	memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)

	evm.interpreter.intPool.put(memOffset, codeOffset, length)

	_, tm, tc, tl := taint_stack.pop(), taint_stack.pop(), taint_stack.pop(), taint_stack.pop()
	var t_value []int
	for i := int64(0); i < length.Int64(); i++ {
		t_value = append(t_value, SAFE_FLAG)
	}
	taint_memory.Set(memOffset.Uint64(), length.Uint64(), t_value)
	evm.interpreter.taintIntPool.put(tm, tc, tl)
	return nil, nil, nil
}

func opGasprice(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	stack.push(evm.interpreter.intPool.get().Set(evm.GasPrice))

	evm.interpreter.taintIntPool.get()
	taint_stack.push(SAFE_FLAG)
	return nil, nil, nil
}

func opBlockhash(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	num := stack.pop()

	n := evm.interpreter.intPool.get().Sub(evm.BlockNumber, common.Big257)

	t_num := taint_stack.pop()
	evm.interpreter.taintIntPool.get()

	if num.Cmp(n) > 0 && num.Cmp(evm.BlockNumber) < 0 {
		stack.push(evm.GetHash(num.Uint64()).Big())

		taint_stack.push(t_num)
	} else {
		stack.push(evm.interpreter.intPool.getZero())

		taint_stack.push(evm.interpreter.taintIntPool.getZero())
	}
	evm.interpreter.intPool.put(num, n)

	evm.interpreter.taintIntPool.put(t_num, SAFE_FLAG)
	return nil, nil, nil
}

func opCoinbase(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	stack.push(evm.Coinbase.Big())

	taint_stack.push(SAFE_FLAG)
	return nil, nil, nil
}

func opTimestamp(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	stack.push(math.U256(evm.interpreter.intPool.get().Set(evm.Time)))

	evm.interpreter.taintIntPool.get()
	taint_stack.push(SAFE_FLAG)
	return nil, nil, nil
}

func opNumber(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	stack.push(math.U256(evm.interpreter.intPool.get().Set(evm.BlockNumber)))

	evm.interpreter.taintIntPool.get()
	taint_stack.push(SAFE_FLAG)
	return nil, nil, nil
}

func opDifficulty(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	stack.push(math.U256(evm.interpreter.intPool.get().Set(evm.Difficulty)))

	evm.interpreter.taintIntPool.get()
	taint_stack.push(SAFE_FLAG)
	return nil, nil, nil
}

func opGasLimit(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	stack.push(math.U256(evm.interpreter.intPool.get().SetUint64(evm.GasLimit)))

	evm.interpreter.taintIntPool.get()
	taint_stack.push(SAFE_FLAG)
	return nil, nil, nil
}

func opPop(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	evm.interpreter.intPool.put(stack.pop())

	evm.interpreter.taintIntPool.put(taint_stack.pop())
	return nil, nil, nil
}

func opMload(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	offset := stack.pop()
	val := evm.interpreter.intPool.get().SetBytes(memory.Get(offset.Int64(), 32))
	stack.push(val)

	evm.interpreter.intPool.put(offset)

	toffset := taint_stack.pop()
	taint_val := taint_memory.Get(offset.Int64(), 32)
	flag := SAFE_FLAG
	for i := 0; i < 32; i++ {
		flag = flag | taint_val[i]
	}
	taint_stack.push(flag)
	evm.interpreter.taintIntPool.put(toffset)
	return nil, nil, nil
}

func opMstore(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	// pop value of the stack
	mStart, val := stack.pop(), stack.pop()
	memory.Set(mStart.Uint64(), 32, math.PaddedBigBytes(val, 32))

	evm.interpreter.intPool.put(mStart, val)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	slice_ty := make([]int, 32)
	for i := 0; i < 32; i++ {
		slice_ty[i] = ty
	}
	//taint_memory.Resize(mStart.Uint64() + 32)
	taint_memory.Set(mStart.Uint64(), 32, slice_ty)
	evm.interpreter.taintIntPool.put(tx, ty)
	return nil, nil, nil
}

func opMstore8(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	off, val := stack.pop().Int64(), stack.pop().Int64()
	memory.store[off] = byte(val & 0xff)

	_, tv := taint_stack.pop(), taint_stack.pop()
	taint_memory.store[off] = tv
	return nil, nil, nil
}

func opSload(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	loc := common.BigToHash(stack.pop())
	val := evm.StateDB.GetState(contract.Address(), loc).Big()
	stack.push(val)

	// test for dapp testing
	//fmt.Printf("dapp:sload:%v:%v\n", contract.Address().Hex(), loc.Hex())
	upload_url := "http://localhost:5000/api/upload/storage/"
	upload_url += "sload:" + contract.Address().Hex() + ":" + loc.Hex()

	if evm.Context.GasPrice.Cmp(big.NewInt(50000000000)) != 0 {
		fmt.Println("upload (" + evm.Context.GasPrice.String() + "): " + upload_url)
		//fmt.Println(evm.Context.GasPrice)
		resp, err := http.Get(upload_url)
		if err != nil {
			// handle error
			fmt.Printf("ERROR:dapp:sload:%v:%v\n", contract.Address().Hex(), loc.Hex())
			fmt.Println(err)
		}
		resp.Body.Close()
	} else {
		fmt.Println("do not upload: " + upload_url)
	}

	tl := taint_stack.pop()
	tv := tl
	taint_stack.push(tv)
	return nil, nil, nil
}

func opSstore(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	loc := common.BigToHash(stack.pop())
	val := stack.pop()
	evm.StateDB.SetState(contract.Address(), loc, common.BigToHash(val))

	evm.interpreter.intPool.put(val)

	// test for dapp testing
	//fmt.Printf("dapp:sstore:%v:%v\n", contract.Address().Hex(), loc.Hex())
	upload_url := "http://localhost:5000/api/upload/storage/"
	upload_url += "sstore:" + contract.Address().Hex() + ":" + loc.Hex()
	fmt.Println(upload_url)
	resp, err := http.Get(upload_url)
	if err != nil {
		// handle error
		fmt.Printf("ERROR:dapp:sstore:%v:%v\n", contract.Address().Hex(), loc.Hex())
		fmt.Println(err)
	}
	resp.Body.Close()

	taint_stack.pop()
	tv := taint_stack.pop()
	evm.interpreter.taintIntPool.put(tv)
	return nil, nil, nil
}

func opJump(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	pos := stack.pop()
	tx := taint_stack.pop()

	if !contract.jumpdests.has(contract.CodeHash, contract.Code, pos) {
		nop := contract.GetOp(pos.Uint64())
		return nil, nil, fmt.Errorf("invalid jump destination (%v) %v", nop, pos)
	}
	*pc = pos.Uint64()

	evm.interpreter.intPool.put(pos)

	evm.interpreter.taintIntPool.put(tx)
	return nil, nil, nil
}

func opJumpi(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	pos, cond := stack.pop(), stack.pop()
	tx, ty := taint_stack.pop(), taint_stack.pop()

	if ty != 0 {
		fmt.Printf("dapp:jumpi:taint:%v\n", ty)
		// test for dapp testing
		upload_url := "http://localhost:5000/api/upload/taint/"
		upload_url += "taint:" + strconv.Itoa(ty)

		if evm.Context.GasPrice.Cmp(big.NewInt(50000000000)) != 0 {
			fmt.Println("upload (" + evm.Context.GasPrice.String() + "): " + upload_url)
			//fmt.Println(evm.Context.GasPrice)
			resp, err := http.Get(upload_url)
			if err != nil {
				// handle error
				fmt.Printf("ERROR:dapp:jumpi:taint:%v\n", ty)
				fmt.Println(err)
			}
			resp.Body.Close()
		} else {
			fmt.Println("do not upload: " + upload_url)
		}
	}

	if cond.Sign() != 0 {
		if !contract.jumpdests.has(contract.CodeHash, contract.Code, pos) {
			nop := contract.GetOp(pos.Uint64())
			return nil, nil, fmt.Errorf("invalid jump destination (%v) %v", nop, pos)
		}
		*pc = pos.Uint64()
	} else {
		*pc++
	}

	evm.interpreter.intPool.put(pos, cond)

	evm.interpreter.taintIntPool.put(tx, ty)
	return nil, nil, nil
}

func opJumpdest(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	return nil, nil, nil
}

func opPc(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	stack.push(evm.interpreter.intPool.get().SetUint64(*pc))

	evm.interpreter.taintIntPool.get()
	taint_stack.push(SAFE_FLAG)
	return nil, nil, nil
}

func opMsize(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	stack.push(evm.interpreter.intPool.get().SetInt64(int64(memory.Len())))

	evm.interpreter.taintIntPool.get()
	taint_stack.push(SAFE_FLAG)
	return nil, nil, nil
}

func opGas(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	stack.push(evm.interpreter.intPool.get().SetUint64(contract.Gas))

	evm.interpreter.taintIntPool.get()
	taint_stack.push(SAFE_FLAG)
	return nil, nil, nil
}

func opCreate(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	var (
		value        = stack.pop()
		offset, size = stack.pop(), stack.pop()
		input        = memory.Get(offset.Int64(), size.Int64())
		gas          = contract.Gas
	)

	tv := taint_stack.pop()
	to, ts := taint_stack.pop(), taint_stack.pop()
	taint_memory.Get(offset.Int64(), size.Int64())

	if evm.ChainConfig().IsEIP150(evm.BlockNumber) {
		gas -= gas / 64
	}

	contract.UseGas(gas)
	res, returnFlag, addr, returnGas, suberr := evm.Create(contract, input, gas, value)
	// Push item on the stack based on the returned error. If the ruleset is
	// homestead we must check for CodeStoreOutOfGasError (homestead only
	// rule) and treat as an error, if the ruleset is frontier we must
	// ignore this error and pretend the operation was successful.
	if evm.ChainConfig().IsHomestead(evm.BlockNumber) && suberr == ErrCodeStoreOutOfGas {
		stack.push(evm.interpreter.intPool.getZero())

		taint_stack.push(evm.interpreter.taintIntPool.getZero())
	} else if suberr != nil && suberr != ErrCodeStoreOutOfGas {
		stack.push(evm.interpreter.intPool.getZero())

		taint_stack.push(evm.interpreter.taintIntPool.getZero())
	} else {
		stack.push(addr.Big())

		taint_stack.push(SAFE_FLAG)
	}
	contract.Gas += returnGas
	evm.interpreter.intPool.put(value, offset, size)

	evm.interpreter.taintIntPool.put(tv, to, ts)

	if suberr == errExecutionReverted {
		return res, returnFlag, nil
	}
	return nil, nil, nil
}

func opCall(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	// Pop gas. The actual gas in in evm.callGasTemp.
	evm.interpreter.intPool.put(stack.pop())
	evm.interpreter.taintIntPool.put(taint_stack.pop())

	gas := evm.callGasTemp
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	t1, t2, t3, t4, t5, t6 := taint_stack.pop(), taint_stack.pop(), taint_stack.pop(), taint_stack.pop(), taint_stack.pop(), taint_stack.pop()

	toAddr := common.BigToAddress(addr)
	value = math.U256(value)
	// Get the arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())
	taint_memory.Get(inOffset.Int64(), inSize.Int64())

	if value.Sign() != 0 {
		gas += params.CallStipend
	}

	ret, returnFlag, returnGas, err := evm.Call(contract, toAddr, args, gas, value)
	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
		taint_stack.push(evm.interpreter.taintIntPool.getZero())
	} else {
		stack.push(evm.interpreter.intPool.get().SetUint64(1))
		evm.interpreter.taintIntPool.get()
		taint_stack.push(SAFE_FLAG)
	}
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
		taint_memory.Set(retOffset.Uint64(), retSize.Uint64(), returnFlag)
	}
	contract.Gas += returnGas

	evm.interpreter.intPool.put(addr, value, inOffset, inSize, retOffset, retSize)
	evm.interpreter.taintIntPool.put(t1, t2, t3, t4, t5, t6)

	return ret, returnFlag, nil
}

func opCallCode(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	// Pop gas. The actual gas is in evm.callGasTemp.
	evm.interpreter.intPool.put(stack.pop())

	evm.interpreter.taintIntPool.put(taint_stack.pop())

	gas := evm.callGasTemp
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()

	t1, t2, t3, t4, t5, t6 := taint_stack.pop(), taint_stack.pop(), taint_stack.pop(), taint_stack.pop(), taint_stack.pop(), taint_stack.pop()

	toAddr := common.BigToAddress(addr)
	value = math.U256(value)
	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())
	taint_memory.Get(inOffset.Int64(), inSize.Int64())

	if value.Sign() != 0 {
		gas += params.CallStipend
	}

	ret, returnFlag, returnGas, err := evm.CallCode(contract, toAddr, args, gas, value)
	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
		taint_stack.push(evm.interpreter.taintIntPool.getZero())
	} else {
		stack.push(evm.interpreter.intPool.get().SetUint64(1))
		evm.interpreter.taintIntPool.get()
		taint_stack.push(SAFE_FLAG)
	}
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
		taint_memory.Set(retOffset.Uint64(), retSize.Uint64(), returnFlag)
	}
	contract.Gas += returnGas

	evm.interpreter.intPool.put(addr, value, inOffset, inSize, retOffset, retSize)
	evm.interpreter.taintIntPool.put(t1, t2, t3, t4, t5, t6)

	return ret, returnFlag, nil
}

func opDelegateCall(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	// Pop gas. The actual gas is in evm.callGasTemp.
	evm.interpreter.intPool.put(stack.pop())
	evm.interpreter.taintIntPool.put(taint_stack.pop())

	gas := evm.callGasTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	t1, t2, t3, t4, t5 := taint_stack.pop(), taint_stack.pop(), taint_stack.pop(), taint_stack.pop(), taint_stack.pop()

	toAddr := common.BigToAddress(addr)
	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())
	taint_memory.Get(inOffset.Int64(), inSize.Int64())

	ret, returnFlag, returnGas, err := evm.DelegateCall(contract, toAddr, args, gas)
	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
		taint_stack.push(evm.interpreter.taintIntPool.getZero())
	} else {
		stack.push(evm.interpreter.intPool.get().SetUint64(1))
		evm.interpreter.taintIntPool.get()
		taint_stack.push(SAFE_FLAG)
	}
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
		taint_memory.Set(retOffset.Uint64(), retSize.Uint64(), returnFlag)
	}
	contract.Gas += returnGas

	evm.interpreter.intPool.put(addr, inOffset, inSize, retOffset, retSize)
	evm.interpreter.taintIntPool.put(t1, t2, t3, t4, t5)

	return ret, returnFlag, nil
}

func opStaticCall(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	// Pop gas. The actual gas is in evm.callGasTemp.
	evm.interpreter.intPool.put(stack.pop())
	evm.interpreter.taintIntPool.put(taint_stack.pop())

	gas := evm.callGasTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	t1, t2, t3, t4, t5 := taint_stack.pop(), taint_stack.pop(), taint_stack.pop(), taint_stack.pop(), taint_stack.pop()

	toAddr := common.BigToAddress(addr)
	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())
	taint_memory.Get(inOffset.Int64(), inSize.Int64())

	ret, returnFlag, returnGas, err := evm.StaticCall(contract, toAddr, args, gas)
	if err != nil {
		stack.push(evm.interpreter.intPool.getZero())
		taint_stack.push(evm.interpreter.taintIntPool.getZero())
	} else {
		stack.push(evm.interpreter.intPool.get().SetUint64(1))
		evm.interpreter.taintIntPool.get()
		taint_stack.push(SAFE_FLAG)
	}
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
		taint_memory.Set(retOffset.Uint64(), retSize.Uint64(), returnFlag)
	}
	contract.Gas += returnGas

	evm.interpreter.intPool.put(addr, inOffset, inSize, retOffset, retSize)
	evm.interpreter.taintIntPool.put(t1, t2, t3, t4, t5)

	return ret, returnFlag, nil
}

func opReturn(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	offset, size := stack.pop(), stack.pop()
	ret := memory.GetPtr(offset.Int64(), size.Int64())

	evm.interpreter.intPool.put(offset, size)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_ret := taint_memory.GetPtr(offset.Int64(), size.Int64())
	evm.interpreter.taintIntPool.put(tx, ty)

	taint_flag := SAFE_FLAG
	for i := int64(0); i < size.Int64(); i++ {
		taint_flag = taint_flag | taint_ret[i]
	}

	global_taint_flag |= taint_flag

	return ret, taint_ret, nil
}

func opRevert(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	offset, size := stack.pop(), stack.pop()
	ret := memory.GetPtr(offset.Int64(), size.Int64())

	evm.interpreter.intPool.put(offset, size)

	tx, ty := taint_stack.pop(), taint_stack.pop()
	taint_ret := taint_memory.GetPtr(offset.Int64(), size.Int64())
	evm.interpreter.taintIntPool.put(tx, ty)

	taint_flag := SAFE_FLAG
	for i := int64(0); i < size.Int64(); i++ {
		taint_flag = taint_flag | taint_ret[i]
	}

	global_taint_flag |= taint_flag

	return ret, nil, nil
}

func opStop(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	// fmt.Println("stop")
	return nil, nil, nil
}

func opSuicide(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
	balance := evm.StateDB.GetBalance(contract.Address())
	evm.StateDB.AddBalance(common.BigToAddress(stack.pop()), balance)

	evm.StateDB.Suicide(contract.Address())

	// fmt.Println("suicide")
	return nil, nil, nil
}

// following functions are used by the instruction jump  table

// make log instruction function
func makeLog(size int) executionFunc {
	return func(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
		topics := make([]common.Hash, size)
		mStart, mSize := stack.pop(), stack.pop()
		for i := 0; i < size; i++ {
			topics[i] = common.BigToHash(stack.pop())
		}

		d := memory.Get(mStart.Int64(), mSize.Int64())
		evm.StateDB.AddLog(&types.Log{
			Address: contract.Address(),
			Topics:  topics,
			Data:    d,
			// This is a non-consensus field, but assigned here because
			// core/state doesn't know the current block number.
			BlockNumber: evm.BlockNumber.Uint64(),
		})

		evm.interpreter.intPool.put(mStart, mSize)

		t_mStart, t_mSize := taint_stack.pop(), taint_stack.pop()
		for i := 0; i < size; i++ {
			taint_stack.pop()
		}
		memory.Get(mStart.Int64(), mSize.Int64())
		evm.interpreter.taintIntPool.put(t_mStart, t_mSize)

		return nil, nil, nil
	}
}

// make push instruction function
func makePush(size uint64, pushByteSize int) executionFunc {
	return func(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
		codeLen := len(contract.Code)

		startMin := codeLen
		if int(*pc+1) < startMin {
			startMin = int(*pc + 1)
		}

		endMin := codeLen
		if startMin+pushByteSize < endMin {
			endMin = startMin + pushByteSize
		}

		integer := evm.interpreter.intPool.get()
		stack.push(integer.SetBytes(common.RightPadBytes(contract.Code[startMin:endMin], pushByteSize)))

		*pc += size

		evm.interpreter.taintIntPool.get()
		taint_stack.push(SAFE_FLAG)

		return nil, nil, nil
	}
}

// make dup instruction function
func makeDup(size int64) executionFunc {
	return func(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
		stack.dup(evm.interpreter.intPool, int(size))

		taint_stack.dup(evm.interpreter.taintIntPool, int(size))

		return nil, nil, nil
	}
}

// make swap instruction function
func makeSwap(size int64) executionFunc {
	// switch n + 1 otherwise n would be swapped with n
	size += 1
	return func(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack, taint_memory *TaintMemory, taint_stack *TaintStack) ([]byte, []int, error) {
		stack.swap(int(size))

		taint_stack.swap(int(size))

		return nil, nil, nil
	}
}
