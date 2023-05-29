import sqlite3
import uuid
from typing import List

DATABASE = 'clients.db'

api_keys: List[str] = [
    "a5f4f9c5ac7047fda7466c6e6c704c4a",
    "17edfaa572154f15a3c6c67ed22d8ebd",
    "5c74b6c67079468cb6348d2d84b26ac5",
    "30d759c34f36419bbaa8da12548c63c9",
    "8b527a17e1ec4110b75568c67a8b7d2b",
    "88baf357c3cf4fcaa0f4d76c78acbf3b",
    "07db4c6d6e6e47f1a42c8b1e1f37a14b",
    "4756c5c884744c9fb5d6e6a0f0c71d9b",
    "10d31a60c0f744f5aa0fdaf60c3c4f6a",
    "9c1bd7d9a5d849c5b0adfdaf6c71d9a5"
]

def create_connection(): 
    return sqlite3.connect(DATABASE)

def close_connection(conn):
    conn.close()

def create_clients_table(conn):
    try:
        sql_create_clients_table = """ CREATE TABLE IF NOT EXISTS clients (
                                            id text PRIMARY KEY,
                                            api_key text NOT NULL
                                        ); """
        if conn is not None:
            c = conn.cursor()
            c.execute(sql_create_clients_table)
    except Exception as e:
        print(e)

def create_valid_api_keys_table(conn):
    try:
        sql_create_api_keys_table = """
        CREATE TABLE IF NOT EXISTS valid_api_keys (
            key TEXT PRIMARY KEY NOT NULL
        );
        """
        conn.execute(sql_create_api_keys_table)
    except Exception as e:
        print(f"An error occurred: {e}")

def is_valid_key(conn: sqlite3.Connection, api_key: str):
    cursor = conn.cursor()
    cursor.execute("SELECT * FROM valid_api_keys WHERE key=?", (api_key,))
    return cursor.fetchone() is not None

def is_key_assigned(conn, api_key):
    cursor = conn.cursor()
    cursor.execute("SELECT * FROM clients WHERE api_key=?", (api_key,))
    return cursor.fetchone() is not None

def assign_api_key_to_client(conn: sqlite3.Connection, client_id: str, api_key: str):
    cursor = conn.cursor()
    cursor.execute("INSERT INTO clients(id, api_key) VALUES (?, ?)", (client_id, api_key))
    conn.commit()

def generate_client_id():
    return str(uuid.uuid4())

def select_client_by_api_key(conn, api_key):
    cursor = conn.cursor()
    cursor.execute("SELECT id FROM clients WHERE api_key=?", (api_key,))
    result = cursor.fetchone()
    return None if result is None else result[0]


def insert_api_keys(conn):
    cursor = conn.cursor()
    for key in api_keys:
        cursor.execute("INSERT OR IGNORE INTO valid_api_keys(key) VALUES (?)", (key,))
    conn.commit()