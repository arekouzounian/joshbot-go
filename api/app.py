from flask import Flask, request 
import csv 
import os 

app = Flask(__name__)

'''
userID,12,0 // userID, total joshes, total non-joshes
userID,1,1 
...
'''
userTable = './users.csv' # stores user info: userID, number of joshes sent

'''
timestamp,userID,1 // user sent a josh on this timestamp
timestamp,userID,0 // user sent a non-josh on this timestamp
...
'''
joshTable = './joshlog.csv' # basically a log of every josh sent 

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
    
    print('asdf')
    

    with open(joshTable, mode='a') as joshlog: 
        writer = csv.writer(joshlog)
        writer.writerow([json['unixTimestamp'], json['userID'], json['joshInt']])

    joshCount = 0 
    nonJoshCount = 0
    if os.path.exists(userTable):
        with open(userTable, mode='r') as user_table_read: 
            reader = csv.reader(user_table_read)
            for row in reader: 
                if row[0] == json['userID']: 
                    joshCount = int(row[1])
                    nonJoshCount = int(row[2])
                    break 
    
    with open(userTable, mode='w') as user_table_write:
        writer = csv.writer(user_table_write)
        if json['joshInt'] == 1: 
            joshCount += 1
        else:
            nonJoshCount += 1 

        writer.writerow([json['userID'], joshCount, nonJoshCount])
    return 'Request success', 200

# whenever a user is added to the server 
# if a user leaves the server their josh count will remain 
# if they rejoin it won't be reset 
@app.route("/api/v1/joshjoin", methods=['POST'])
def memberUpdate(): 
    '''
    {
    "userID": "12345" // their discord userID
    "totalUsers": ["userID", "userID", ...] // full list of server userID's 
    }
    '''

    json = request.get_json()
    requiredFields = ['userID', 'totalUsers']
    missingFields = checkFields(requiredFields, json)
    
    if len(missingFields) > 0:
        return missingFieldsErr(missingFields)


    # check if file exists, gather all users in table 
    existing = []
    if os.path.exists(userTable):
        with open(userTable, mode='r') as csv_file: 
            reader = csv.reader(csv_file)
            for row in reader: 
                existing.append(row[0])

    # figure out which users we need to add 
    to_add = []
    if json['userID'] not in existing: 
        to_add.append(json['userID'])
    for user in json['totalUsers']: 
        if user not in existing: 
            to_add.append(user)

    # add all outstanding users to table
    with open(userTable, mode='a') as csv_file: 
        writer = csv.writer(csv_file)
        for user in to_add: 
            writer.writerow([json['userID'], '0', '0'])
    
    return 'User successfully joined.'
                



