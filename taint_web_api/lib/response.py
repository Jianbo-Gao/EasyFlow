#!/usr/bin/env python2
# -*- coding: utf-8 -*-

from flask import Response
import json



def success(msg, color, title):
    response={}
    response["code"]=0
    response["status"]="success"
    response["data"]=str(msg)
    response["color"]=str(color)
    response["title"]=str(title)
    return Response(json.dumps(response),mimetype="application/json")

def fail(msg, color, title):
    response={}
    response["code"]=1
    response["status"]="fail"
    response["data"]=str(msg)
    response["color"]=str(color)
    response["title"]=str(title)
    return Response(json.dumps(response),mimetype="application/json")
