package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type FileState struct {
	LastModified time.Time
	FileName     string
}

const address = "localhost:6060"

func main() {

	apiKey := os.Args[1]           // Read API key from command line arguments
	directoryToWatch := os.Args[2] //

	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Send the API key to server with delimiter
	conn.Write([]byte(fmt.Sprintf("%s\n", apiKey)))

	message, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Print("Message from server: " + message)

	if message == "Failed to Authenticate\n" {
		fmt.Println("Failed to authenticate")
		return
	}

	conn.Write([]byte("OK\n"))

	fileSystemEventsChannel := make(chan string, 100)

	// receive the inital set of updates before cofiguring watch
	go receiveFileSyncUpdates(&conn, directoryToWatch, fileSystemEventsChannel)

	go watchFileSysUpdates(&conn, directoryToWatch, fileSystemEventsChannel)

	select {} // run forever
}

func watchFileSysUpdates(conn *net.Conn, directoryToWatch string, eventsChannel chan string) {
	watcher, err := NewWatcher()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = watcher.WatchDirectories(directoryToWatch)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer watcher.Close()

	for {
		action, type_, name, size, err := watcher.Next()
		if err != nil {
			fmt.Println(err)
			break
		}

		// if name is empty then skip sending update
		if name == "" {
			continue
		}
		name = strings.TrimPrefix(name, directoryToWatch)
		//actionStr := fmt.Sprintf("ACTION : %s, TYPE : %s , NAME : %s , SIZE : %s", action, type_, name, size)

		// select {
		// case update := <-eventsChannel:
		// 	fmt.Println("Update " + update)
		// 	if update == actionStr {
		// 		fmt.Printf("[Ignore] - Same as incoming event %s", update)
		// 		continue
		// 	}
		// default:
		// 	fmt.Println("Sleep")
		// 	time.Sleep(2 * time.Second)
		// }
		sendFile(conn, action, type_, name, size)
	}
}

func sendFile(conn *net.Conn, action, type_, name, size string) {
	actionStr := fmt.Sprintf("ACTION : %s, TYPE : %s , NAME : %s , SIZE : %s\n", action, type_, name, size)

	_, err := (*conn).Write([]byte(actionStr))
	if err != nil {
		fmt.Println("Error sending action string:", err)
		return
	}

	response, err := bufio.NewReader(*conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error receiving response:", err)
		return
	}
	response = strings.TrimSpace(response)

	switch type_ {
	case "DIR":
		if response == "OK" {
			return
		}
	case "FILE":
		if response == "READY" {
			// Open the file
			file, err := os.Open(name)
			if err != nil {
				fmt.Println("Error reading file:", err)
				return
			}

			_, err = io.Copy((*conn), file)
			if err != nil {
				fmt.Println("Error sending file:", err)
				return
			}

			response, err = bufio.NewReader(*conn).ReadString('\n')
			if err != nil {
				fmt.Println("Error receiving response:", err)
				return
			}

			response = strings.TrimSpace(response)
			if response == "OK" {
				return
			}
		}
	default:
		fmt.Printf("No action taken for type %s\n", type_)
	}

	_, err = (*conn).Write([]byte("DONE\n"))
	if err != nil {
		fmt.Println("Error sending DONE:", err)
		return
	}
}

func receiveFileSyncUpdates(conn *net.Conn, directoryToWatch string, eventsChannel chan string) {
	connection := *conn
	reader := bufio.NewReader(connection)
	fmt.Println("Syncing with server")

	for {
		line, err := reader.ReadString(byte('\n')) // use '\n' as delimiter
		fmt.Println("[Received] - " + line)
		if err != nil {
			fmt.Println("Error reading:", err)
			return // consider if we want to return or take some other action here
		}

		// if we encounter done that means streaming is done
		if strings.Trim(line, "\n") == "DONE" {
			continue
		}
		// fmt.Println("Sending line")
		// // send message into events channel as well

		// select {
		// case eventsChannel <- line:
		// 	fmt.Println("message sent")

		// default:
		// 	fmt.Println("message not sent")
		// }

		action, type_, name, size, err := parseActionString(line)
		if err != nil {
			fmt.Println(err)
		}

		_, err = os.Stat(directoryToWatch)
		if os.IsNotExist(err) {
			os.MkdirAll(directoryToWatch, 0777)
		}
		os.Chdir(directoryToWatch)
		os.Chmod(directoryToWatch, 0777)

		err = processActionString(action, type_, name, size, conn)
		if err != nil {
			fmt.Println("Error processing action string : " + err.Error())
		}
	}
}

func parseActionString(actionStr string) (string, string, string, string, error) {
	parts := strings.Split(actionStr, ",")

	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("invalid action string")
	}

	actionPart := strings.Split(strings.TrimSpace(parts[0]), " : ")
	if len(actionPart) != 2 {
		return "", "", "", "", fmt.Errorf("invalid action string")
	}

	typePart := strings.Split(strings.TrimSpace(parts[1]), " : ")
	if len(typePart) != 2 {
		return "", "", "", "", fmt.Errorf("invalid type string")
	}

	namePart := strings.Split(strings.TrimSpace(parts[2]), " : ")
	if len(namePart) != 2 {
		return "", "", "", "", fmt.Errorf("invalid name string")
	}

	sizePart := strings.Split(strings.TrimSpace(parts[3]), " : ")
	if len(sizePart) != 2 {
		return "", "", "", "", fmt.Errorf("invalid size string")
	}

	return strings.TrimSpace(actionPart[1]), strings.TrimSpace(typePart[1]), strings.TrimSpace(namePart[1]), strings.TrimSpace(sizePart[1]), nil
}

func processActionString(action string, objectType string, objectName string, objectSize string, tcpConn *net.Conn) error {
	connection := *tcpConn

	if action == "CREATE" {

		switch strings.Trim(objectType, " ") {
		case "FILE":
			// derive total number of bytes we need to read
			bytesToRead, err := strconv.ParseInt(strings.Trim(objectSize, "\n"), 10, 64)
			if err != nil {
				return err
			}

			// Open a file for writing
			file, err := os.Create(objectName)
			if err != nil {
				return err
			}

			err = os.Chmod(objectName, 0777)
			if err != nil {
				return err
			}

			defer file.Close()

			connection.Write([]byte("READY\n")) // send readiness confirmation after processing the line

			// Stream data from the connection to the file
			_, err = io.CopyN(file, connection, bytesToRead)
			if err != nil {
				return err
			}

			fmt.Printf("File Created %s\n", objectName)
		case "DIR":
			// Create a directory
			err := os.MkdirAll(objectName, os.ModePerm)
			if err != nil {
				return err
			}

			fmt.Printf("Dir Created %s\n", objectName)
		default:
			fmt.Printf("No action taken for type %s\n", objectType)
		}

		connection.Write([]byte("OK\n")) // send done confirmation after processing the line
	}

	return nil
}
