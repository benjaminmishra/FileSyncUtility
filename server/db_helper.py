import sqlite3

DATABASE = 'clients.db'

def create_connection(): 
    return sqlite3.connect(DATABASE)

def close_connection(conn):
    conn.close()

def create_table(conn):
    try:
        sql_create_clients_table = """ CREATE TABLE IF NOT EXISTS clients (
                                            id integer PRIMARY KEY,
                                            api_key text NOT NULL,
                                            jwt_token text
                                        ); """
        if conn is not None:
            c = conn.cursor()
            c.execute(sql_create_clients_table)
    except Exception as e:
        print(e)

def insert_client(conn:sqlite3.Connection, api_key:str, generated_jwt_token:str):
    sql = ''' INSERT INTO clients(api_key,jwt_token)
              VALUES(?,?) '''
    cur = conn.cursor()
    cur.execute(sql,[api_key, generated_jwt_token])
    conn.commit()
    return cur.lastrowid

def select_client_by_api_key(conn:sqlite3.Connection, api_key:str)->str:
    cur = conn.cursor()
    cur.execute("SELECT * FROM clients WHERE api_key=?", (api_key,))

    rows = cur.fetchall()

    if len(rows)>0:
        return rows[0]["jwt_token"]
    else:
        return ""

def update_client_jwt_token(conn, api_key, jwt_token):
    sql = ''' UPDATE clients
              SET jwt_token = ?
              WHERE api_key = ?'''
    cur = conn.cursor()
    cur.execute(sql, (jwt_token, api_key,))
    conn.commit()
