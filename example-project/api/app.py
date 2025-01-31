# app.py
from flask import Flask, request, jsonify
import redis
import json

app = Flask(__name__)
r = redis.Redis(host='redis', port=6379, db=0)

@app.route('/api/send', methods=['POST'])
def send_data():
    payload = request.json
    r.publish('hooks', json.dumps(payload))
    return jsonify({"status": "sent to channel"})

if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0')
