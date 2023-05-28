import threading
import socket
from dirsync import sync_dir, read_until_newline, process_line
import os

class ClientThread(threading.Thread):
    """ Handles client threads """
    threads = {}

    def __init__(self, client_address:str, client_socket:socket.socket, api_key:str):
        threading.Thread.__init__(self)
        self.csocket = client_socket
        self.client_addr = client_address
        self.api_key = api_key
        print("New connection added: ", client_address)

    def run_inital_download(self):
        print("Connection from : ", self.client_addr)

        for action, type, name, size in sync_dir("../monitor"):
            action_str = f"ACTION : {action}, TYPE : {type} , NAME : {name}, SIZE : {size}\n"
            
            self.csocket.sendall(bytes(action_str,"utf-8"))
            data = self.csocket.recv(1024).decode()
            
            if type=="DIR":    
                if data == "OK":
                    continue
            
            if type=="FILE":
                if data=="READY":
                    with open(name,'rb') as file:
                        print(f"Sending file {name}...\n")
                        self.csocket.sendall(file.read())
                        response = self.csocket.recv(1024).decode()
                        if response == "OK":
                            continue
    
        self.csocket.sendall(bytes("DONE\n","utf-8"))


    def listen_to_updates(self):
        print("Listning at : ", self.client_addr)

        while True:
            update_action_line = read_until_newline(self.csocket)

            process_line(self.csocket, update_action_line)
            


