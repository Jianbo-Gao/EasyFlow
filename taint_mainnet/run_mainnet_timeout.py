#!/usr/bin/env python
# -*- coding: utf-8 -*-

import pymysql
import os, sys, json

import jsonrpc

import subprocess, datetime, time, signal


def TIMEOUT_COMMAND(command, timeout):
    cmd = command.split(" ")
    start = datetime.datetime.now()
    process = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    while process.poll() is None:
        time.sleep(0.2)
        now = datetime.datetime.now()
        if (now - start).seconds> timeout:
            os.kill(process.pid, signal.SIGKILL)
            os.waitpid(-1, os.WNOHANG)
            return None
    return process.stdout.read()


MYSQL_HOST="localhost"
MYSQL_USER="root"
MYSQL_PASS="password"
MYSQL_DB="blockchain_sync"

JSONRPC_HOST="http://192.168.1.44:8545"

EVM_PATH = "/root/real-mainnet/EasyFlow/build/bin/evm"
TIMEOUT_SECONDS = 60

SELECT_SQL="select c_id,c_from,c_contract_addr,c_input_data from T_ethereum_contract_small where easyflow_result is NULL and length(c_input_data)<1000 limit 1000"
UPDATE_SQL="update T_ethereum_contract_small SET easyflow_result=%s, easyflow_lastop=%s WHERE c_id=%s"

conn = pymysql.connect(host=MYSQL_HOST,user=MYSQL_USER,passwd=MYSQL_PASS,db=MYSQL_DB)
cur = conn.cursor()

while True:
    cur.execute(SELECT_SQL)
    print(datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')+" "+"select success")
    sys.stdout.flush()
    contracts = cur.fetchall()
    print(datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')+" "+"fetch success")
    sys.stdout.flush()

    if len(contracts) == 0:
        break

    for contract in contracts:
        c_id = contract[0]
        c_from = contract[1]
        c_contract_addr = contract[2]
        c_input_data = contract[3]

        evm_code = jsonrpc.getcode(JSONRPC_HOST,c_contract_addr)
        evm_input = c_input_data[2:]
        evm_from = c_from
        evm_to = c_contract_addr

        print(datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')+" "+"detecting "+str(c_id))
        sys.stdout.flush()

        #output = os.popen("%s --code %s --input %s --sender %s --receiver %s --json run" % (EVM_PATH, evm_code, evm_input, evm_from, evm_to)).read()
        #print(output)
        cmd = "%s --code %s --input %s --sender %s --receiver %s --json run" % (EVM_PATH, evm_code, evm_input, evm_from, evm_to)
        output = TIMEOUT_COMMAND(cmd, TIMEOUT_SECONDS)

        if not output:
            print(datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')+" "+"timeout: kill "+str(c_id))
            cur.execute(UPDATE_SQL, ("TIMEOUT","TIMEOUT",c_id))
            conn.commit()
            continue

        try:
            evm_lastop = json.loads(output.splitlines()[-3])["opName"]
        except:
            try:
                evm_lastop = json.loads(output.splitlines()[-6])["opName"]
            except:
                evm_lastop = "JSON-ERROR"

        try:
            evm_result = output.splitlines()[-1].strip().split(":")[1][1:]
        except:
            evm_result = "JSON-ERROR"

        cur.execute(UPDATE_SQL, (evm_result,evm_lastop,c_id))

        print(datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')+" "+str(c_id)+","+evm_result+","+evm_lastop)
        sys.stdout.flush()
        conn.commit()

        if c_id % 100 == 0:
            conn.commit()

    conn.commit()


conn.commit()
cur.close()
conn.close()
