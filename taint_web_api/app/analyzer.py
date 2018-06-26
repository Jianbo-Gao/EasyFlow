#!/usr/bin/env python2
# -*- coding: utf-8 -*-
import flask_restful as restful
from flask_restful import reqparse
import sys, json, requests, os, uuid


sys.path.append("..")
from conf import * 
from lib import *

class Analyzer(restful.Resource):
    def __init__(self):
        self.output=""

    def post(self):
        try:
            parser = reqparse.RequestParser()
            parser.add_argument("type", type=unicode, required=True)
            parser.add_argument("code", type=unicode, required=True)
            parser.add_argument("input", type=unicode, required=True)
            param = parser.parse_args()
            if param["type"] == "bytecode":
                self.main(param["code"], param["input"])
            elif param["type"] == "solidity":
                status, compile_result = self.compile_solidity(param["code"])
                if status:
                    self.main(compile_result, param["input"])
                else:
                    return response.fail(compile_result)
            return response.success(self.output)
        except Exception, e:
            return response.fail(str(e))

    def compile_solidity(self, solidity_code):
        try:
            solidity_id = str(uuid.uuid1())
            solidity_filepath = os.path.join("/tmp",solidity_id+".sol")
            solidity_resultdir = os.path.join("/tmp",solidity_id)
            with open(solidity_filepath, 'w') as f:
                f.write(solidity_code)
            os.popen("%s --bin-runtime -o %s %s" % (common.SOLC_PATH, solidity_resultdir, solidity_filepath))
            compile_result = os.popen("cat %s" % os.path.join(solidity_resultdir,"*.bin-runtime")).read()
            return True, compile_result
        except Exception, e:
            return False, str(e)

    def print_res(self, id, input_str, last_op, taint_res):
        self.output += "[Tx %s]\n" % id
        self.output += "input: %s\n" % input_str
        self.output += "result: %s\n" % taint_res
        self.output += "\n"

    def run_evm(self, code_str, input_str):
        output = os.popen("%s --code %s --input %s --json run" % (common.EVM_PATH, code_str, input_str))
        output_str = output.read()
        try:
            last_op = json.loads(output_str.splitlines()[-3])["opName"]
        except:
            last_op = json.loads(output_str.splitlines()[-6])["opName"]
        taint_res = output_str.splitlines()[-1].strip().split(":")[1][1:]
        return last_op, taint_res

    def run_evm_with_value(self, code_str, input_str):
        output = os.popen("%s --code %s --input %s --sender 0000000000000000000000000000000000000000 --value \"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff\" --prestate %s --json run" % (common.EVM_PATH, code_str, input_str, common.PRESTATE_PATH))
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

    def main(self, code_str, input_str):
        last_op, taint_res = self.run_evm(code_str, input_str)
        self.print_res(0, input_str, last_op, taint_res)
        if taint_res in ("safe", "overflow", "protected overflow"):
            self.output += taint_res+"\n"
            if taint_res == "overflow":
                return True, last_op, taint_res, None
            else:
                return False, last_op, taint_res, None

        elif taint_res == "potential overflow":
            last_op_with_value, taint_res_with_value = self.run_evm_with_value(code_str, input_str)
            self.print_res("0 with value", input_str, last_op_with_value, taint_res_with_value)
            if taint_res_with_value == "overflow":
                retry_result = "retry: potential overflow triggered"
                self.output += retry_result+"\n"
                return True, last_op_with_value, retry_result, None

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
                retry_last_op, retry_taint_res = self.run_evm(code_str, input_str)
                self.print_res(retry_id, input_str, retry_last_op, retry_taint_res)
                if retry_taint_res == "overflow":
                    retry_result = "retry: potential overflow triggered"
                    self.output += retry_result + "\n"
                    return True, retry_last_op, retry_result, None
            else:
                retry_result = "retry: potential overflow NOT triggered"
                self.output += retry_result +"\n"
                return False, last_op, retry_result, None

        else:
            self.output += taint_res + "\n"
            self.output += "ERROR\n"
            return False, last_op, taint_res, "ERROR"
