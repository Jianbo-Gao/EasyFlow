#!/usr/bin/env python2
# -*- coding: utf-8 -*-

from flask import Response
import json



def success(msg=""):
    response={}
    response["code"]=0
    response["status"]="success"
    response["data"]=msg
    return Response(json.dumps(response),mimetype="application/json")

def fail(msg=""):
    response={}
    response["code"]=1
    response["status"]="fail"
    response["data"]=msg
    return Response(json.dumps(response),mimetype="application/json")