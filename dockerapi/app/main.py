from flask import Flask, request, jsonify
from flask_cors import CORS
from apscheduler.schedulers.background import BackgroundScheduler
from apscheduler.triggers.cron import CronTrigger
import sqlite3
import time 
import random
import atexit

app = Flask(__name__)
CORS(app)

DB_LOC = '/db/josh.db'
LEADERBOARD_COUNT = 5

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
    except:
        rollback = True 
        return 'Internal Server Error', 500
    finally:
        if not rollback:
            conn.commit()
        else:
            conn.rollback()
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
    except:
        rollback = True
        return 'Internal Server Error', 500
    finally:
        if not rollback:
            conn.commit()
        else:
            conn.rollback()
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
    except:
        return 'Internal Server Error', 500
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
        
    except:
        return 'Internal Server Error', 500
    finally:
        conn.close()

@app.route('/api/v2/joshcount/<userid>')
def getJoshes(userid):
    try: 
        conn = sqlite3.connect(DB_LOC)
        curs = conn.cursor()
        (josh_count) = curs.execute(f"SELECT COUNT(*) FROM joshlog WHERE user_id={userid} AND is_josh=1").fetchone()
        if josh_count is None: 
            return 'Internal Server Error', 500
        
        (non_josh_count) = curs.execute(f"SELECT COUNT(*) FROM joshlog WHERE user_id={userid} AND is_josh=0").fetchone()
        if josh_count is None:
            return 'Internal Server Error', 500

        return jsonify([josh_count[0], non_josh_count[0]]), 200
        
    except:
        return 'Internal Server Error', 500
    finally:
        conn.close()
        
@app.route('/api/v2/joshboard')
def getLeaderboard():
    try:
        conn = sqlite3.connect(DB_LOC)
        curs = conn.cursor()
        # select user_id, count(*) as count_occurrences from joshlog where is_josh = 1 group by user_id order by count_occurrences desc limit 5;
        rows = curs.execute(f'SELECT user_id, COUNT(*) AS count_occurrences FROM joshlog WHERE is_josh=1 GROUP BY user_id ORDER BY count_occurrences DESC LIMIT 5').fetchall()
        
        # should be [userID, username, avatarURL, joshCount, nonJoshCount]
        # we can leave nonJoshCount field blank
        for i in range(len(rows)):
            [userID, joshes] = rows[i]
            res = curs.execute(f'SELECT * FROM users WHERE user_id = {userID}').fetchone()
            if res is None or len(res) < 4:
                return 'Internal Server Error', 500
            (_, username, avatar_url, _) = res

            rows[i] = [userID, username, avatar_url, joshes, 0]
        
        return jsonify(rows), 200
    except Exception as e:
        return 'Internal Server Error', 500
    finally:
        conn.close()

@app.route('/api/v2/joshofshame')
def getWallOfShame():
    try:
        conn = sqlite3.connect(DB_LOC)
        curs = conn.cursor()
        # select user_id, count(*) as count_occurrences from joshlog where is_josh = 1 group by user_id order by count_occurrences desc limit 5;
        rows = curs.execute(f'SELECT user_id, COUNT(*) AS count_occurrences FROM joshlog WHERE is_josh=0 GROUP BY user_id ORDER BY count_occurrences DESC LIMIT 5').fetchall()
        # should be [userID, username, avatarURL, joshCount, nonJoshCount]
        # we can leave nonJoshCount field blank
        for i in range(len(rows)):
            [userID, joshes] = rows[i]
            print(userID, joshes)

            res = curs.execute(f'SELECT * FROM users WHERE user_id={userID}').fetchone()
            if res is None or len(res) < 4:
                continue
            (_, username, avatar_url, _) = res

            rows[i] = [userID, username, avatar_url, joshes, 0]
        
        return jsonify(rows), 200
    except Exception as e:
        return 'Internal Server Error', 500
    finally:
        conn.close()


@app.route('/api/v2/joshotw')
def getJoshOTW():
    try:
        conn = sqlite3.connect(DB_LOC)
        curs = conn.cursor()
        row = curs.execute('SELECT * FROM users WHERE josh_otw=2').fetchone()
        if row is None:
            joshOfTheWeek() # create new one
            row = curs.execute('SELECT * FROM users WHERE josh_otw=2').fetchone()
        
        return jsonify(row), 200
    except:
        return 'Internal Server Error', 500
    finally:
        conn.close()


def joshOfTheWeek():
    rollback = False

    # 0 if you haven't been josh of the week
    # 1 if you've previously been josh of the week 
    # 2 if you're currently josh of the week 

    try:
        conn = sqlite3.connect(DB_LOC)
        curs = conn.cursor()
        rows = curs.execute(f'SELECT * FROM users WHERE josh_otw=0').fetchall()

        # everyone has been one, full reset 
        if len(rows) < 1:
            curs.execute(f'UPDATE users SET josh_otw=0 WHERE josh_otw=2 OR josh_otw=1')
            rows = curs.execute(f'SELECT * FROM users WHERE josh_otw=0').fetchall()
        
        (user_id, _, _, _) = random.choice(rows)

        # find curr josh, set to 1
        curs.execute(f'UPDATE users SET josh_otw=1 WHERE josh_otw=2')
        curs.execute(f'UPDATE users SET josh_otw=2 WHERE user_id={user_id}')

            
    except:
        rollback = True
    finally:
        if not rollback:
            conn.commit()
        else:
            conn.rollback()

        conn.close()

scheduler = BackgroundScheduler()
everyWeekTrigger = CronTrigger(year='*',month='*',week='*',day_of_week='0',hour='0',minute='0',second='0')
scheduler.add_job(joshOfTheWeek, trigger=everyWeekTrigger)
scheduler.start()

atexit.register(scheduler.shutdown)




# if __name__ == "__main__":
#     # Only for debugging while developing
#     app.run(host='0.0.0.0', debug=True, port=80)
#     app.run()