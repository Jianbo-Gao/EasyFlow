#!/usr/bin/env python
# -*- coding: utf-8 -*-

import os, json, argparse
import config

def print_res(id, input_str, last_op, taint_res):
    print("[Tx %s]" % id)
    print("input: %s" % input_str)
    #print("last op: %s" % last_op)
    print("result: %s" % taint_res)
    print("")

def run_evm(code_str, input_str):
    output = os.popen("%s --code %s --input %s --json run" % (config.EVM_PATH, code_str, input_str))
    output_str = output.read()
    try:
        last_op = json.loads(output_str.splitlines()[-3])["opName"]
    except:
        last_op = json.loads(output_str.splitlines()[-6])["opName"]
    taint_res = output_str.splitlines()[-1].strip().split(":")[1][1:]
    return last_op, taint_res

def run_evm_with_value(code_str, input_str):
    output = os.popen("%s --code %s --input %s --sender 0000000000000000000000000000000000000000 --value \"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff\" --prestate genesis-example.json --json run" % (config.EVM_PATH, code_str, input_str))
    output_str = output.read()
    line = -3
    last_op=None
    while line >= -10:
        try:
            last_op = json.loads(output_str.splitlines()[line])["opName"]
            break
        except:
            line -= 1
    taint_res = output_str.splitlines()[-1].strip().split(":")[1][1:]
    return last_op, taint_res

def main(code_str, input_str, debug_flag=False):
    last_op, taint_res = run_evm(code_str, input_str)
    debug_flag and print_res(0, input_str, last_op, taint_res)
    if taint_res in ("safe", "overflow", "protected overflow"):
        #print(last_op)
        print(taint_res)
        if taint_res == "overflow":
            return True, last_op, taint_res, None
        else:
            return False, last_op, taint_res, None

    elif taint_res == "potential overflow":
        last_op_with_value, taint_res_with_value = run_evm_with_value(code_str, input_str)
        debug_flag and print_res("0 with value", input_str, last_op_with_value, taint_res_with_value)
        if taint_res_with_value == "overflow":
            #print(last_op_with_value)
            print("retry: potential overflow triggered")
            return True, last_op_with_value, taint_res_with_value, None

        arg_num = (len(input_str)-8)/64
        input_args = ("0000000000000000000000000000000000000000000000000000000000000000", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
        input_method = input_str[0:8]

        retry_sum = 2 ** arg_num
        retry_id = 0
        while retry_id < retry_sum:
            input_str = input_method
            arg_id = 0
            while arg_id < arg_num:
                input_str += input_args[(retry_id>>arg_id)%2]
                arg_id += 1

            retry_id += 1
            retry_last_op, retry_taint_res = run_evm(code_str, input_str)
            debug_flag and print_res(retry_id, input_str, retry_last_op, retry_taint_res)
            if retry_taint_res == "overflow":
                #print(retry_last_op)
                retry_result = "retry: potential overflow triggered"
                print(retry_result)
                return True, retry_last_op, retry_result, None
        else:
            #print(last_op)
            retry_result = "retry: potential overflow NOT triggered"
            print(retry_result)
            return False, last_op, retry_result, None

    else:
        #print(last_op)
        print(taint_res)
        print("ERROR")
        return False, last_op, taint_res, "ERROR"

def init_parser(help=False):
    parser = argparse.ArgumentParser(description='Overflow Analyzer for EVM')
    parser.add_argument('--demo', action="store_true", help='use demo code and input')
    parser.add_argument('-d', '--debug', action="store_true", help='print debug info')
    parser.add_argument('-c', '--code', help='runtime bytecode')
    parser.add_argument('-i', '--input', help='input data')
    return parser

def demo():
    test_code = "608060405260043610610062576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806333a581d214610067578063710419ab14610092578063b79e70ed146100dd578063c1b9e80314610128575b600080fd5b34801561007357600080fd5b5061007c610173565b6040518082815260200191505060405180910390f35b34801561009e57600080fd5b506100c76004803603810190808035906020019092919080359060200190929190505050610197565b6040518082815260200191505060405180910390f35b3480156100e957600080fd5b5061011260048036038101908080359060200190929190803590602001909291905050506101d3565b6040518082815260200191505060405180910390f35b34801561013457600080fd5b5061015d60048036038101908080359060200190929190803590602001909291905050506101f7565b6040518082815260200191505060405180910390f35b7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff81565b6000817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff038311156101c857600080fd5b818301905092915050565b6000828284011015156101ed5781830192508290506101f1565b8290505b92915050565b600080828401905083811015151561020b57fe5b80915050929150505600a165627a7a72305820fd40ac0f90592eaf3991dc1806889e89f3b6e2447ce5f5e8a2e0f25f3cc648910029e"
    test_input = "b79e70edffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0000000000000000000000000000000000000000000000000000000000000002"

    print("###DEBUG###")
    print("code: %s" % test_code)
    print("input: %s" % test_input)
    print("")

    main(test_code, test_input)

if __name__ == '__main__':

    parser = init_parser()
    args = parser.parse_args()

    if args.demo and not (args.code or args.input):
        demo()
    elif (not args.demo) and args.code and args.input:
        main(args.code, args.input, args.debug)
    else:
        parser.print_help()

