import socket
from dirsync import sync_dir, read_until_newline, process_line
from constants import ROOT_DIR
import os
import db_helper
import broadcaster as b


class ClientHandler:
    """ Handles clients """

    def __init__(self, client_address: str, client_socket: socket.socket, clinet_id: str, broadcaster: b.Broadcaster):
        self.csocket = client_socket
        self.client_addr = client_address
        self.db_conn = db_helper.create_connection()
        self.__broadcaster = broadcaster
        self.__broadcaster.register(clinet_id, self)
        self._clinet_id = clinet_id
        print("New connection added: ", client_address)

    def handle(self):
        self.__send_initial_state()
        self.__listen_to_updates_and_broadcast()
        return

    def __send_initial_state(self):
        print("Connection from : ", self.client_addr)

        os.chdir(ROOT_DIR)

        for action, type_, name, size in sync_dir(ROOT_DIR):
            self.send_update(action, type_, name, size)

        self.csocket.sendall(bytes("DONE\n", "utf-8"))

    def send_update(self, action: str, type_: str, name: str, size: int):
        action_str = f"ACTION : {action}, TYPE : {type_} , NAME : {name}, SIZE : {size}\n"
        print(f"[Sending] - {action_str}")

        self.csocket.sendall(bytes(action_str, "utf-8"))
        data = read_until_newline(self.csocket)

        if data is None:
            self.__shutdown()
            return

        if type == "DIR":
            if data == "OK":
                return

        if type == "FILE":
            if data == "READY":
                with open(name, 'rb') as file:
                    print(f"Sending file {name}...\n")
                    self.csocket.sendall(file.read())
                    response = read_until_newline(self.csocket)
                    if response is None:
                        self.__shutdown()
                        return

                    if response == "OK":
                        return

    def __listen_to_updates_and_broadcast(self):
        while True:
            print(f"Listening for updates from : {self.client_addr}")
            update_action_line = read_until_newline(self.csocket)

            if not update_action_line:
                self.__shutdown()
                return

            if update_action_line == "DONE":
                continue

            action_details = process_line(self.csocket, update_action_line)
            if action_details is None:
                print(f"Could not process update file system action {update_action_line}")
                return

            action, type_, name, size = action_details
            self.__broadcaster.broadcast(self._clinet_id, action, type_, name, size)

    def __shutdown(self):
        self.csocket.close()
        print("Client disconnected, shtting down thread")
        return

    def client_is_connected(self):
        try:
            # send a dummy message to test the connection
            self.csocket.send(bytes("", 'UTF-8'))
            return True
        except socket.error:
            return False
