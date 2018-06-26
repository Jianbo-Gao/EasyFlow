#!/usr/bin/env python2
# -*- coding: utf-8 -*-

import sys  
sys.path.append("/usr/local/lib/python2.7/dist-packages")

from flask import Flask
import flask_restful as restful

from conf import *
from app import *


app = Flask(__name__)
api = restful.Api(app)


api.add_resource(analyzer.Analyzer, "/api/analyze")


if __name__ == '__main__':
    app.debug = True
    app.run(host='0.0.0.0')


