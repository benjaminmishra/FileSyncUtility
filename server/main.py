"""Main entry point of the application"""

from concurrent.futures import ThreadPoolExecutor
import socket
from clienthandler import ClientHandler
from constants import LOCALHOST, PORT, SECRET_KEY
import db_helper
from typing import Dict, Tuple
from dirsync import read_until_newline

# dict to keep track of clients
connected_clients: Dict[str, ClientHandler] = dict()


def autheticate_client(api_key:str) -> Tuple[str, bool]:
    client_guid = ""
    is_success = False
    if db_helper.is_valid_key(db_conn, api_key):
        client_guid = db_helper.select_client_by_api_key(db_conn, api_key)
        # new client 
        if client_guid is None:
            client_guid = db_helper.generate_client_id()
            db_helper.assign_api_key_to_client(db_conn, client_guid, api_key)
            print(f"Assigned API Key {api_key} to Client {client_guid}!")

        is_success = True

    return client_guid, is_success  


if __name__ == "__main__":
    db_conn = db_helper.create_connection()
    db_helper.create_valid_api_keys_table(db_conn)
    db_helper.create_clients_table(db_conn)
    db_helper.insert_api_keys(db_conn)
    
    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server.setsockopt(socket.SOL_SOCKET, socket.SO_KEEPALIVE, 1)
    server.bind((LOCALHOST, PORT))
    
    print("Server started")
    
    server.listen(32)

    with ThreadPoolExecutor(max_workers=32) as executor:
        while True:
            print(f"Listning on port {PORT} for new clients")
            client_sock, client_addr = server.accept()
            api_key = read_until_newline(client_sock)

            client_id,success = autheticate_client(api_key)

            authecation_message = "SUCCESS\n"

            if not success:
                print(f"API Key {api_key} is not valid!")
                authecation_message = "FAILURE\n"

            client_sock.sendall(bytes(authecation_message,"utf-8"))

            client_handler = ClientHandler(client_addr, client_sock)
            connected_clients[client_id] = client_handler
            executor.submit(client_handler.handle)


        