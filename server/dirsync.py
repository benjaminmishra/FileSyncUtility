import os
from socket import socket

def sync_dir(root_dir):
    for dirpath, _ , filenames in os.walk(root_dir):
        
        yield "CREATE", "DIR" , dirpath , 0

        for file_name in filenames:
            full_file_name = os.path.join(dirpath, file_name)
            size = os.path.getsize(full_file_name)

            yield "CREATE", "FILE", os.path.join(dirpath, file_name), size
        
        