from flask import Flask, request, jsonify
from flask_cors import CORS
import sqlite3
import time 

app = Flask(__name__)
CORS(app)

DB_LOC = '/db/josh.db'

# TODO: get rid of string errors and just use 'Internal Server Error'

def checkFields(fields_arr, json):
    missing = []
    for field in fields_arr:
        if field not in json.keys():
            missing.append(field)

    return missing

def missingFieldsErr(missing):
    msg = 'Missing the following fields: '
    for field in missing:
        msg += f'{field},'

    return msg

@app.route("/")
def hello():
    return "josh"

@app.route("/test")
def sql_test():
    conn = sqlite3.connect(DB_LOC)
    curs = conn.cursor()

    obj = [row for row in curs.execute('SELECT * FROM users')]

    conn.close()

    return jsonify(obj)

@app.route('/api/v2/newjosh', methods=['POST'])
def newJosh():
    '''
    {
    "userID": "12345" // discord userID of the person who sent the josh
    "unixTimestamp": 12345 // unix-time timestamp of when the message was sent
    "joshInt": 1 // 1 if the message was 'josh', 0 otherwise
    }
    '''
    json = request.get_json()
    requiredFields = ['userID', 'unixTimestamp', 'joshInt']
    missingFields = checkFields(requiredFields, json)
    if len(missingFields) > 0:
        return missingFieldsErr(missingFields)
    
    rollback = False
    
    # insert into the joshlog 
    try:
        conn = sqlite3.connect(DB_LOC)
        curs = conn.cursor()
        curs.execute(f"INSERT INTO joshlog VALUES({json['unixTimestamp']}, {json['userID']}, {json['joshInt']})")
    except Exception as e:
        rollback = True 
        return str(e), 500
    finally:
        if not rollback:
            conn.commit()
        conn.close()
    

    return 'Request sucess', 200

@app.route('/api/v2/joshupdate', methods=['POST'])
def memberUpdate():
    '''
    {
    "userID": "12345" // their discord userID
    "username: "ikigag" // their discord username
    "avatar": "http://cdn.discord.com/pfp.png"
    }
    '''
    json = request.get_json()
    requiredFields = ['userID', 'username', 'avatar']
    missingFields = checkFields(requiredFields, json)
    if len(missingFields) > 0:
        return missingFieldsErr(missingFields)
    
    rollback = False
    
    try: 
        conn = sqlite3.connect(DB_LOC)
        curs = conn.cursor()
        curs.execute(f"UPDATE users SET user_id='{json['userID']}',username='{json['username']}',avatar_url='{json['avatar']}' WHERE user_id={json['userID']}")
    except Exception as e:
        rollback = True
        return str(e), 500
    finally:
        if not rollback:
            conn.commit()
        conn.close()
    
    return 'Request success',200


@app.route('/api/v2/lastjosh', methods=['GET'])
def timeSinceLastJosh():
    now = int(time.time())
    
    try:
        conn = sqlite3.connect(DB_LOC)
        curs = conn.cursor()
        res = curs.execute('SELECT timestamp FROM joshlog WHERE ROWID IN ( SELECT max( ROWID ) FROM joshlog )').fetchall()

        if len(res) < 1:
            return 'Internal Server Error', 500
        
        return str(now - int(res[0][0])), 200
    except Exception as e:
        return str(e), 500
    finally:
        conn.close()

@app.route('/api/v2/joshavg', methods=['GET'])
def joshAvg():
    josh_avg = 0
    non_josh_avg = 0
    
    thirty_days_ago = int(time.time()) - (30 * 24 * 3600)
    try:
        conn = sqlite3.connect(DB_LOC)
        curs = conn.cursor()
        res = curs.execute(f'SELECT COUNT(*) FROM joshlog WHERE timestamp >= {thirty_days_ago} AND is_josh=1').fetchone()
        if res is None: 
            return 'Internal Server Error', 500
        
        josh_avg = res[0] / 30 

        res = curs.execute(f'SELECT COUNT(*) FROM joshlog WHERE timestamp >= {thirty_days_ago} AND is_josh=0').fetchone()
        if res is None:
            return 'Internal Server Error', 500
        
        non_josh_avg = res[0] / 30 

        return jsonify([josh_avg, non_josh_avg]), 200
        
    except Exception as e:
        return str(e), 500
    finally:
        conn.close()



# if __name__ == "__main__":
#     # Only for debugging while developing
#     app.run(host='0.0.0.0', debug=True, port=80)
#     app.run()