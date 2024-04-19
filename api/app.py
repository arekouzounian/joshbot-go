from flask import Flask, request, jsonify 
from flask_cors import CORS
from apscheduler.schedulers.background import BackgroundScheduler
from apscheduler.triggers.interval import IntervalTrigger
import csv 
import os 
import time
import random
import atexit 


app = Flask(__name__)   
CORS(app)


'''
userID,ikigag,http://mypfp.url,12,0 // userID, username, avatar url, total joshes, total non-joshes
userID,ikigag2,http://otherpfp.url,1,1
...
'''
USER_TABLE_ID_OFFSET = 0
USER_TABLE_USERNAME_OFFSET = 1
USER_TABLE_AVATAR_OFFSET = 2
USER_TABLE_JOSH_OFFSET = 3
USER_TABLE_NONJOSH_OFFSET = 4
userTable = './users.csv' # stores user info: userID, number of joshes sent
USER_TABLE_NUM_FIELDS = 5 # number of fields (columns) in the table 


'''
timestamp,userID,1 // user sent a josh on this timestamp
timestamp,userID,0 // user sent a non-josh on this timestamp
...
'''
JOSH_TABLE_TIMESTAMP_OFFSET = 0
JOSH_TABLE_ID_OFFSET = 1
JOSH_TABLE_JOSHINT_OFFSET = 2
joshTable = './joshlog.csv' # basically a log of every josh sent 
JOSH_TABLE_NUM_FIELDS = 3 # number of fields (columns) in the table 

JOSH_OTW_TABLE_ID_OFFSET = 0 
JOSH_OTW_TABLE_USERNAME_OFFSET = 1
JOSH_OTW_TABLE_AVATAR_OFFSET = 2 
JOSH_OTW_TABLE_JOSH_OFFSET = 3 
JOSH_OTW_TABLE_NONJOSH_OFFSET = 4
joshOfTheWeekTable = './joshOTW.csv'
JOSH_OTW_TABLE_NUM_FIELDS = 5


# slow and hacky but it's all python anyway right 
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
def test(): 
    return '<h1>josh</h1>'


#TODO: update joshlog to be in reverse chronological order 

# new josh message 
@app.route("/api/v1/newjosh", methods=['POST'])
def newJosh():     
    '''
    {
    "userID": "12345" // discord userID of the person who sent the josh 
    " Timestamp": 12345 // unix-time timestamp of when the message was sent
    "joshInt": 1 // 1 if the message was 'josh', 0 otherwise
    }
    '''
    json = request.get_json()
    requiredFields = ['userID', 'unixTimestamp', 'joshInt']
    missingFields = checkFields(requiredFields, json)
    if len(missingFields) > 0:
        return missingFieldsErr(missingFields)
    
    with open(joshTable, mode='a') as joshlog: 
        writer = csv.writer(joshlog)
        writer.writerow([json['unixTimestamp'], json['userID'], json['joshInt']])

    userRow = -1
    table = []
    if os.path.exists(userTable):
        with open(userTable, mode='r') as user_table_read: 
            reader = csv.reader(user_table_read)
            table = list(reader)
            for (i, row) in enumerate(table): 
                if row[USER_TABLE_ID_OFFSET] == json['userID']: 
                    userRow = i 
                    break
            if userRow < 0: 
                return "Error: user doesn't exist", 500
    else: 
        return "Error: user table doesn't exist", 500
    
    with open(userTable, mode='w') as user_table_write:
        writer = csv.writer(user_table_write)
        if json['joshInt'] == 1: 
            table[userRow][USER_TABLE_JOSH_OFFSET] = (int)(table[userRow][USER_TABLE_JOSH_OFFSET]) + 1
        else:
            table[userRow][USER_TABLE_NONJOSH_OFFSET] = (int)(table[userRow][USER_TABLE_NONJOSH_OFFSET]) + 1

        writer.writerows(table)

    
    return 'Request success', 200

# whenever a user is added to the server 
# if a user leaves the server their josh count will remain 
# if they rejoin it won't be reset 
@app.route("/api/v1/joshupdate", methods=['POST'])
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


    userRow = -1
    table = []
    if os.path.exists(userTable): 
        with open(userTable, mode='r') as csv_file: 
            reader = csv.reader(csv_file)
            table = list(reader)
            for (i, row) in enumerate(table): 
                if row[USER_TABLE_ID_OFFSET] == json['userID']:
                    userRow = i

    # add default fields if new user 
    if userRow < 0:
        table.append(['0'] * USER_TABLE_NUM_FIELDS)
        table[userRow][USER_TABLE_JOSH_OFFSET] = 0
        table[userRow][USER_TABLE_NONJOSH_OFFSET] = 0
    # update relevant info 
    table[userRow][USER_TABLE_ID_OFFSET] = json['userID']
    table[userRow][USER_TABLE_USERNAME_OFFSET] = json['username']
    table[userRow][USER_TABLE_AVATAR_OFFSET] = json['avatar']

    with open(userTable, mode='w') as csv_file: 
        writer = csv.writer(csv_file)
        writer.writerows(table)
    
    return 'User successfully joined.'
                


#TODO: update this to only grab first line (given that it's stored in reverse chrono order)

# number of seconds since last josh 
@app.route("/api/v1/lastjosh", methods=['GET'])
def timeSinceLastJosh():
    # check log for last josh 

    if os.path.exists(joshTable):
        timestamp = 0
        with open(joshTable, mode='r') as csv_file: 
            reader = csv.reader(csv_file)
            for line in reader: 
                pass
            timestamp = line[JOSH_TABLE_TIMESTAMP_OFFSET]

        if timestamp == 0: 
            return 'Error reading table',500
        
        return str(int(time.time()) - int(timestamp)), 200
    else:
        return 'Josh Table doesn\'t exist', 500



# check average joshes per day over last 30 days 
@app.route("/api/v1/joshavg", methods=['GET'])
def joshAvg(): 
    if not os.path.exists(joshTable):
        return 'Josh Table doesn\'t exist', 500 
    
    josh_avg = 0 
    non_josh_avg = 0

    thirty_days_ago = int(time.time()) - (30 * 24 * 3600)
    with open(joshTable, mode='r') as csv_file: 
        reader = csv.reader(csv_file)
        for row in reader: 
            if int(row[JOSH_TABLE_TIMESTAMP_OFFSET]) < thirty_days_ago:
                continue 

            if row[JOSH_TABLE_JOSHINT_OFFSET] == '0': 
                non_josh_avg += 1
            else: 
                josh_avg += 1 
    
    josh_avg /= 30 
    non_josh_avg /= 30 

    return jsonify([josh_avg, non_josh_avg]), 200 

# get number of joshes, nonjoshes for an individual user 
@app.route("/api/v1/joshcount/<userid>")
def getJoshes(userid): 
    if not os.path.exists(userTable):
        return 'User table doesn\'t exist!', 500
    
    with open(userTable, 'r') as csv_file: 
        reader = csv.reader(csv_file)
        for row in reader:
            if row[USER_TABLE_ID_OFFSET] == userid:
                return jsonify([row[USER_TABLE_JOSH_OFFSET], row[USER_TABLE_NONJOSH_OFFSET]]), 200
    
    return 'User doesn\'t exist!', 500


# board_len = number of entries in the leaderboard
# sort_key = lambda expression for how the entries should be sorted 
def getBoard(board_len, sort_key): 
    if not os.path.exists(userTable):
        return 'User table doesn\'t exist!', 500
    
    with open(userTable, 'r') as c: 
        reader = csv.reader(c)
        users = list(reader)
        users.sort(key=sort_key, reverse=True)
        return jsonify(users if len(users) < board_len else users[:board_len]), 200
    

# get top 5 users by josh count 
@app.route("/api/v1/joshboard")
def getLeaderboard(): 
    return getBoard(5, lambda x: int(x[USER_TABLE_JOSH_OFFSET]))
    
# get top 5 users by non josh count 
@app.route("/api/v1/joshofshame")
def getWallOfShame(): 
    return getBoard(5, lambda x: int(x[USER_TABLE_NONJOSH_OFFSET]))


@app.route("/api/v1/joshotw")
def getJoshOTW():
    if not os.path.exists(joshOfTheWeekTable): 
        return 'Josh OTW table doesn\'t exist!', 500

    with open(joshOfTheWeekTable, 'r') as c:
        reader = csv.reader(c)
        return jsonify(list(reader)), 200


def joshOfTheWeek(): 
    # pick a new josh 
    if not os.path.exists(userTable): 
        return 'User table doesn\'t exist!', 500
    
    row = []
    with open(userTable, 'r') as c: 
        reader = csv.reader(c)
        lines = list(reader)
        row = random.choice(lines)

        # check for duplicate josh 
        if os.path.exists(joshOfTheWeekTable):
            with open(joshOfTheWeekTable, 'r') as j:
                first_line = next(csv.reader(j))
                # bogosort-like maneuver
                while row == first_line:
                    row = random.choice(lines)



    with open(joshOfTheWeekTable, 'w') as j: 
        line = ''
        for i, item in enumerate(row):
            line += item
            if i != len(row) - 1:
                line += ','

        line += '\n'

        j.write(line)

    return line, 200


scheduler = BackgroundScheduler()
scheduler.add_job(joshOfTheWeek, IntervalTrigger(weeks=1))
scheduler.start()

atexit.register(scheduler.shutdown)