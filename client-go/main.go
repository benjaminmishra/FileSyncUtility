package main

import (
	"bufio"
	"client/jwt"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type FileState struct {
	LastModified time.Time
	FileName     string
}

const address = "localhost:6060"
const secretKey = "mu9vTDxsLDZMfqP9NP+l81WjG6t4yYe8H8gLKH2X9wE="

//var filesState = make(map[string]FileState)

func main() {

	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	conn.Write([]byte("NewConn"))

	message, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Print("Message from server: " + message)

	jwtToken := strings.Trim(message, "\n")

	payload, err := jwt.DecodeAndVerifyPayload(jwtToken, secretKey)
	if err != nil {
		fmt.Println("Invalid JWT")
	}

	fmt.Printf("Payload: %+v\n", payload)

	conn.Write([]byte("OK"))

	w, err := jwt.NewWatcher()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer w.Close()

	initialFileSync(&conn, w)

	watchFiles(w, &conn)

	select {} // run forever
}

func watchFiles(w *jwt.Watcher, conn *net.Conn) {
	for {
		println("Starting file watch")
		action, type_, name, size, err := w.Next()
		if err != nil {
			fmt.Println(err)
			break
		}
		time.Sleep(time.Second)
		println("Sending updated file")
		go sendFile(conn, action, type_, name, size)
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
			fileContent, err := ioutil.ReadFile(name)
			if err != nil {
				fmt.Println("Error reading file:", err)
				return
			}

			_, err = (*conn).Write(fileContent)
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
}

func initialFileSync(conn *net.Conn, watcher *jwt.Watcher) {
	connection := *conn
	reader := bufio.NewReader(connection)

	for {

		line, err := reader.ReadString('\n') // use '\n' as delimiter

		if err != nil {
			fmt.Println("Error reading:", err)
			return // consider if you want to return or take some other action here
		}

		// if we encounter done that means streaming is done
		if strings.TrimRight(line, "\n") == "DONE" {
			break
		}

		action, type_, name, size, err := parseActionString(line)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("Action: %s, Type: %s, Name: %s, Size %s\n", action, type_, name, size)

		if action == "CREATE" {
			new_name := filepath.Base(name)
			os.Chdir("../monitor")

			switch strings.Trim(type_, " ") {
			case "FILE":
				// derive total number of bytes we need to read
				bytesToRead, err := strconv.ParseInt(strings.Trim(size, "\n"), 10, 64)
				if err != nil {
					log.Fatal(err)
				}

				// Open a file for writing
				file, err := os.Create(new_name)
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()

				connection.Write([]byte("READY")) // send readiness confirmation after processing the line

				// Stream data from the connection to the file
				_, err = io.CopyN(file, connection, bytesToRead)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Printf("File Created %s\n", name)
			case "DIR":
				// Create a directory
				err := os.Mkdir(new_name, os.ModePerm)
				if err != nil {
					fmt.Println(err)
				}

				err = (*watcher).WatchDirectories(new_name)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Printf("Dir Created %s\n", name)
			default:
				fmt.Printf("No action taken for type %s\n", type_)
			}

			connection.Write([]byte("OK")) // send done confirmation after processing the line
		}
	}
}

func parseActionString(actionStr string) (string, string, string, string, error) {
	parts := strings.Split(actionStr, ", ")

	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("invalid action string")
	}

	actionPart := strings.Split(parts[0], " : ")
	if len(actionPart) != 2 {
		return "", "", "", "", fmt.Errorf("invalid action string")
	}

	typePart := strings.Split(parts[1], " : ")
	if len(typePart) != 2 {
		return "", "", "", "", fmt.Errorf("invalid type string")
	}

	namePart := strings.Split(parts[2], " : ")
	if len(namePart) != 2 {
		return "", "", "", "", fmt.Errorf("invalid name string")
	}

	sizePart := strings.Split(parts[3], " : ")
	if len(sizePart) != 2 {
		return "", "", "", "", fmt.Errorf("invalid size string")
	}

	return actionPart[1], typePart[1], namePart[1], sizePart[1], nil
}
