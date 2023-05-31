from typing import Dict, Any
import threading

lock = threading.Lock()

class Broadcaster:
    """ Resposible for keeping track of clients that are connected 
        Use the broadcast function to send file updates to every connected client other than the source
    """

    def __init__(self):
        # dict to keep track of clients
        self.connected_clients: Dict[str, Any] = dict()

    def register(self, client_id: str, handler: object):
        """Keeps track of client vs client handler in a underlying dict
           Provides single threaded guarantee
        """
        with lock:
            self.connected_clients[client_id] = handler


    def broadcast(self, source_client_id: str, action: str, type_: str, name: str, size: int):
        # lock to ensure no one else is add to the list at the same time
        with lock:
            # send updates to everyone execpt the source and disconnected ones
            for client_id, handler in self.connected_clients.items():
                if source_client_id != client_id and handler.client_is_connected():
                    handler.send_update(action, type_, name, size);
