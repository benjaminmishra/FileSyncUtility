import os
from socket import socket

def sync_dir(root_dir):
    print(root_dir)
    for dirpath, _ , filenames in os.walk(root_dir):
        if dirpath == root_dir:
            continue 
        yield "CREATE", "DIR" , dirpath , 0

        for file_name in filenames:
            full_file_name = os.path.join(dirpath, file_name)
            size = os.path.getsize(full_file_name)

            yield "CREATE", "FILE", os.path.join(dirpath, file_name), size

def read_until_newline(conn: socket):
    message = ''
    while True:
        chunk = conn.recv(1)  # read one byte
        if chunk == b'\n':   # check if the byte is newline
            break
        message += chunk.decode('utf-8')
    return message

def parse_action_string(action_str):
    parts = action_str.split(",")

    if len(parts) != 4:
        raise ValueError("Invalid action string")

    action_part = parts[0].split(":")
    if len(action_part) != 2:
        raise ValueError("Invalid action string")

    type_part = parts[1].split(":")
    if len(type_part) != 2:
        raise ValueError("Invalid type string")

    name_part = parts[2].split(":")
    if len(name_part) != 2:
        raise ValueError("Invalid name string")

    size_part = parts[3].split(":")
    if len(size_part) != 2:
        raise ValueError("Invalid size string")

    return action_part[1].strip(), type_part[1].strip(), name_part[1].strip(), size_part[1].strip()


def process_line(sock, line):
    action, type_, name, size = parse_action_string(line)

    print(f"Action: {action}, Type: {type_}, Name: {name}, Size {size}")
    
    new_name = os.path.basename(name)
    type_ = type_.strip()
    size = int(size.strip())
    os.chdir("/home/benjaminmishra/monitor")
    if action == "CREATE" or action == "MODIFY":
        if type_ == "FILE":
            # Open a file for writing
            with open(new_name, 'wb') as file:
                sock.sendall(b"READY")  # send readiness confirmation after processing the line
                # Stream data from the connection to the file
                data = sock.recv(size)
                file.write(data)
                print(f"File {'Modified' if action == 'MODIFY' else 'Created'} {name}")

        elif type_ == "DIR" and action == "CREATE":
            # Create a directory
            try:
                os.makedirs(new_name)
                print(f"Dir Created {name}")
            except OSError as err:
                print(err)

        else:
            print(f"No action taken for type {type_}")

        sock.sendall(b"OK")  # send done confirmation after processing the line
