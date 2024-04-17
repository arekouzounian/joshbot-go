from flask import Flask, request 
import csv 
import os 

app = Flask(__name__)

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


# new josh message 
@app.route("/api/v1/newjosh", methods=['POST'])
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
                



