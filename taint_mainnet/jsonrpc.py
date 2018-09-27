import requests
import json

def getcode(jsonrpc_url, contract_addr):
    data  = {"jsonrpc":"2.0","method":"eth_getCode","params":[contract_addr, "latest"],"id":1}
    headers = {"Content-Type": "application/json"}

    req = requests.session()
    response = req.post(jsonrpc_url,data=json.dumps(data),headers=headers)

    return json.loads(response.content)["result"][2:]


if __name__ == '__main__':
    print getcode("http://192.168.1.44:8545", "0x6e38a457c722c6011b2dfa06d49240e797844d66")
