'''
The purpose of this script is to locate legacy joshbot files, 
create sqlite3 tables, and populate them with the legacy joshbot info
'''

import sqlite3
import csv, json

# Creates the following tables: 
#   - users
#   - coins
#   - joshlog
def create_tables(): 
    conn = sqlite3.connect('./josh.db')
    curs = conn.cursor()

    # User table: user_id,username,avatar_url,josh_otw
    # removed josh, non-josh fields because we can calculate that from the josh log 
    # it's technically slower in the large case but I'll save the bridge crossing for when I'm crossing bridges

    curs.execute('CREATE TABLE IF NOT EXISTS users(user_id TEXT, username TEXT, avatar_url TEXT, josh_otw INTEGER)')
    
    # joshlog: timestamp,user_id,is_josh
    curs.execute('CREATE TABLE IF NOT EXISTS joshlog(timestamp INTEGER, user_id TEXT, is_josh INTEGER)')

    # coins: user_id,coins_before,coins_today
    curs.execute('CREATE TABLE IF NOT EXISTS coins(user_id TEXT, coins_before INTEGER, coins_today INTEGER)')

    conn.commit()
    conn.close()

def populate_tables(users_csv_path, log_csv_path, joshcoin_path, josh_otw_path):
    conn = sqlite3.connect('./josh.db')
    curs = conn.cursor()

    josh_otw_id = 0

    with open(josh_otw_path, 'r') as f:
        r = csv.reader(f)
        for line in r: 
            josh_otw_id = line[0]
            
            
    if josh_otw_id == 0: 
        print('No josh of the week!')
        exit(1)

    with open(users_csv_path, 'r') as f: 
        r = csv.reader(f)
        for line in r: 
            [user_id, username, avatar_url, josh, non_josh] = line
            josh_otw = 1 if josh_otw_id == user_id else 0
            curs.execute(f"INSERT INTO users VALUES('{user_id}', '{username}', '{avatar_url}', {josh_otw})")
    

    with open(log_csv_path, 'r') as f:
        r = csv.reader(f)
        for line in r: 
            [timestamp, user_id, is_josh] = line
            curs.execute(f"INSERT INTO joshlog VALUES({timestamp}, '{user_id}', {is_josh})")

    with open(joshcoin_path, 'r') as f:
        obj = json.load(f)
        for id in obj['dailyCoinsEarned']:
            curs.execute(f"INSERT INTO coins VALUES('{id}', {obj['coinsBeforeToday'][id]}, {obj['dailyCoinsEarned'][id]})") 

    conn.commit()
    conn.close()


# just empties out whole db
# dangerous function
def drop_tables():
    conn = sqlite3.connect('./josh.db')
    curs = conn.cursor()

    curs.execute('DROP TABLE IF EXISTS users')
    curs.execute('DROP TABLE IF EXISTS joshlog')
    curs.execute('DROP TABLE IF EXISTS coins')
    
    conn.commit()
    conn.close()
        
    

if __name__ == '__main__':
    drop_tables()
    create_tables()
    populate_tables('../api/users.csv', '../api/joshlog.csv', '../bot/joshcointables.json', '../api/joshOTW.csv')