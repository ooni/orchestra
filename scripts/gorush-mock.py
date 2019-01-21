# -*- coding: utf-8 -*-
import json
from pprint import pprint
from flask import Flask
from flask import request

app = Flask(__name__)

@app.route('/api/push', methods=["GET", "POST", "PUT", "HEAD"])
def api_push():
    output = u"┏" + u"━"*80
    output += "\n"
    output += u"┃ request.method: %s" % request.method
    output += "\n"
    output += u"┃ request.full_path: %s" % request.full_path
    output += "\n"
    output += u"┗" + u"━"*80
    output += "\n"
    output += json.dumps(request.get_json(), indent=2)
    output += "\n"
    output += u"━"*80
    output += "\n"
    print(output)
    return 'scheduled!'

if __name__ == '__main__':
    print("Listening on 127.0.0.1:8081")
    app.run(host='127.0.0.1', port=8081)
