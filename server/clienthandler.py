import threading
import socket
from dirsync import sync_dir, read_until_newline, process_line
from constants import ROOT_DIR
import os
import db_helper

class ClientHandler():
    """ Handles clients """

    def __init__(self, client_address:str, client_socket:socket.socket):
        self.csocket = client_socket
        self.client_addr = client_address
        self.db_conn = db_helper.create_connection()
        print("New connection added: ", client_address)

    def handle(self):
        self.__send_initial_state()
        self.__listen_to_updates()
        return

    def __send_initial_state(self):
        print("Connection from : ", self.client_addr)
        os.chdir(ROOT_DIR)
        for action, type, name, size in sync_dir(ROOT_DIR):
            action_str = f"ACTION : {action}, TYPE : {type} , NAME : {name}, SIZE : {size}\n"
            print(action_str)
            self.csocket.sendall(bytes(action_str,"utf-8"))
            data = read_until_newline(self.csocket)
            if data is None:
                self.__shutdown()
                return
            
            if type=="DIR":    
                if data == "OK":
                    continue
            
            if type=="FILE":
                if data=="READY":
                    with open(name,'rb') as file:
                        print(f"Sending file {name}...\n")
                        self.csocket.sendall(file.read())
                        response = read_until_newline(self.csocket)
                        if response is None:
                            self.__shutdown()
                            return

                        if response == "OK":
                            continue
    
        self.csocket.sendall(bytes("DONE\n","utf-8"))

    def __listen_to_updates(self):
        while True:
            print(f"Listening for updates from : {self.client_addr}")
            update_action_line = read_until_newline(self.csocket)

            if not update_action_line:
                self.__shutdown()
                return

            if update_action_line == "DONE":
                continue;
            
            process_line(self.csocket, update_action_line)

    def __shutdown(self):
        self.csocket.close()
        print("Client disconnected, shtting down thread")
        return

    def __client_is_connected(self):
        try:
            # send a dummy message to test the connection
            self.csocket.send(bytes("", 'UTF-8'))
            return True
        except socket.error:
            return False
        
        
    def __brodcast(self):
        
        
