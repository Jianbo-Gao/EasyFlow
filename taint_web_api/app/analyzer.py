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
        self.color = common.COLOR_GREEN
        self.title = ""
        self.contract_id = str(uuid.uuid1())
        self.contract_dir = os.path.join(common.HISTORY_LOCAL_PATH,self.contract_id)
        os.makedirs(self.contract_dir)

    def post(self):
        #try:
        parser = reqparse.RequestParser()
        parser.add_argument("type", type=unicode, required=True)
        parser.add_argument("name", type=unicode, required=True)
        parser.add_argument("code", type=unicode, required=True)
        parser.add_argument("input", type=unicode, required=True)
        param = parser.parse_args()
        if param["type"] == "bytecode":
            self.main(param["code"], param["input"])
        elif param["type"] == "solidity":
            status, compile_result = self.compile_solidity(param["name"], param["code"])
            if status:
                self.main(compile_result, param["input"])
            else:
                self.color=common.COLOR_GREY
                return response.fail(compile_result, self.color, self.title)
        return response.success(self.output, self.color, self.title)
        #except Exception, e:
        #    return response.fail(str(e), self.color, self.title)

    def compile_solidity(self, contract_name, solidity_code):
        try:
            solidity_filepath = os.path.join("/tmp",self.contract_id+".sol")
            solidity_resultdir = os.path.join("/tmp",self.contract_id)
            with open(solidity_filepath, 'w') as f:
                f.write(solidity_code)
            bin_runtime_filepath=os.path.join(solidity_resultdir, contract_name+".bin-runtime")
            os.popen("%s --bin-runtime -o %s %s" % (common.SOLC_PATH, solidity_resultdir, solidity_filepath))
            if not os.path.isfile(bin_runtime_filepath):
                return False, "Compile Error with solc 0.4.24+commit.e67f0147.Linux.g++."
            compile_result = os.popen("cat %s" % bin_runtime_filepath).read()
            return True, compile_result
        except Exception, e:
            return False, str(e)

    def print_res(self, id, input_str, last_op, taint_res):
        self.output += "<strong>[Transaction %s]</strong>&nbsp;&nbsp;&nbsp;&nbsp;<a class=\"am-btn am-btn-default am-btn-xs am-radius\" target=\"_blank\" href=\"%s\">Show Full Transaction Execution Log</a>\n" % (str(id), common.HISTORY_URL+self.contract_id+"/"+str(id)+".txt")
        self.output += "<strong>input</strong>: %s\n" % input_str
        self.output += "<strong>result</strong>: %s\n" % taint_res
        self.output += "\n"

    def save_detail(self, id, output_str):
        filename = os.path.join(self.contract_dir, str(id)+".txt")
        with open(filename, "w") as f:
            f.write(output_str)

    def run_evm(self, id, code_str, input_str):
        output = os.popen("%s --code %s --input %s --json run" % (common.EVM_PATH, code_str, input_str))
        output_str = output.read()
        self.save_detail(id, output_str)
        try:
            last_op = json.loads(output_str.splitlines()[-3])["opName"]
        except:
            try:
                last_op = json.loads(output_str.splitlines()[-6])["opName"]
            except:
                last_op = ""
        taint_res = output_str.splitlines()[-1].strip().split(":")[1][1:]
        return last_op, taint_res

    def run_evm_with_value(self, id, code_str, input_str):
        output = os.popen("%s --code %s --input %s --sender 0000000000000000000000000000000000000000 --value \"0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff\" --prestate %s --json run" % (common.EVM_PATH, code_str, input_str, common.PRESTATE_PATH))
        output_str = output.read()
        self.save_detail(id, output_str)
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
        last_op, taint_res = self.run_evm(0, code_str, input_str)
        self.print_res(0, input_str, last_op, taint_res)
        if taint_res in ("safe", "overflow", "protected overflow"):
            res_dict = {"safe":"Safe", "overflow": "Overflow", "protected overflow": "Protected Overflow"}
            self.output += res_dict[taint_res]+"\n"
            self.title = res_dict[taint_res]
            if taint_res == "overflow":
                self.color = common.COLOR_RED
                return True, last_op, taint_res, None
            else:
                return False, last_op, taint_res, None

        elif taint_res == "potential overflow":
            last_op_with_value, taint_res_with_value = self.run_evm_with_value("0 with value", code_str, input_str)
            self.print_res("0 with value", input_str, last_op_with_value, taint_res_with_value)
            if taint_res_with_value == "overflow":
                retry_result = "Potential Overflow Triggered"
                self.output += retry_result+"\n"
                self.title = retry_result
                self.color = common.COLOR_RED
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
                retry_last_op, retry_taint_res = self.run_evm(retry_id, code_str, input_str)
                self.print_res(retry_id, input_str, retry_last_op, retry_taint_res)
                if retry_taint_res == "overflow":
                    retry_result = "Potential Overflow Triggered"
                    self.output += retry_result + "\n"
                    self.color = common.COLOR_RED
                    self.title = retry_result
                    return True, retry_last_op, retry_result, None
            else:
                retry_result = "Potential Overflow not Triggered"
                self.output += retry_result +"\n"
                self.title = retry_result
                return False, last_op, retry_result, None

        else:
            self.output += taint_res + "\n"
            self.output += "ERROR\n"
            self.title = "Internal ERROR"
            return False, last_op, taint_res, "ERROR"
