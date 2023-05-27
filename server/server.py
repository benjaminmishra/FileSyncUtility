"""server.py"""


import socket
import secrets
from clientthread import ClientThread
from constants import LOCALHOST, PORT, SECRET_KEY
from db_helper import create_connection, create_table, select_client_by_api_key, insert_client
import jwt
from typing import Dict
from dirsync import sync_dir


client_threads : Dict[str,ClientThread] = {}

if __name__ == "__main__":
    conn = create_connection()
    create_table(conn)
    
    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server.bind((LOCALHOST, PORT))
    
    print("Server started")
    print("Waiting for client request..")

    while True:
        server.listen(32)
        client_sock, client_addr = server.accept()
        jwt_token = client_sock.recv(2048).decode()

        if jwt_token == "NewConn":
            # new client 
            new_api_key = secrets.token_urlsafe(16)
            payload_data = {
                "api_key": new_api_key
            }
            new_jwt_token = jwt.encode(payload_data,SECRET_KEY, algorithm="HS256")
            insert_client(conn,new_api_key,new_jwt_token)

            message = f'{new_jwt_token}\n'

            client_sock.sendall(bytes(message,"utf-8"))
            response = client_sock.recv(1024)

            print(new_jwt_token)
            print(new_api_key)

            new_thread = ClientThread(client_addr, client_sock, new_api_key)
            client_threads[new_api_key] = new_thread

            if(response.decode()=="OK"):
                new_thread.run()

            client_sock.shutdown(1)
            client_sock.close()

        else:
            data = jwt.decode(jwt_token, SECRET_KEY, algorithms=['HS256'])
            jwt_token = select_client_by_api_key(conn,data["api_key"])

            # api key does not match, reject 
            if jwt_token is None:
                client_sock.sendall(bytes("Failed to Autheticate\n","utf-8"))
                client_sock.shutdown(-1)
                client_sock.close()

            client_thread = client_threads[data["api_key"]]
            client_thread.run()

    #close_connection(conn)
