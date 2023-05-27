package main

import (
	"bufio"
	"client/jwt"
	"fmt"
	"io"
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

var filesState = make(map[string]FileState)

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

	initialFileSync(conn)

	w, err := jwt.NewWatcher()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer w.Close()

	err = w.WatchDirectories("../go-client-monitor")
	if err != nil {
		fmt.Println(err)
		return
	}

	go watchFiles(w)

	select {} // run forever
}

func watchFiles(w *jwt.Watcher) {
	for {
		action, type_, name, size, err := w.Next()

		if err != nil {
			fmt.Println(err)
			break
		}

		path := filepath.Join(dir, name)

		fileState, ok := filesState[path]
		if !ok || action == "UPDATE" {
			// If the file is not in our map, or if it was modified,
			// update our map and send the file.
			fileState = FileState{
				LastModified: time.Now(), // we could also use file.ModTime() here
				FileName:     name,
			}
			filesState[path] = fileState
			go sendUpdate(path, fileState)
		}
	}
}

func sendUpdate(path string, fileState FileState) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// listen for reply
	message, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Print("Message from server: " + message)
}

func initialFileSync(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n') // use '\n' as delimiter
		if err != nil {
			fmt.Println("Error reading:", err)
			return // consider if you want to return or take some other action here
		}

		action, type_, name, size, err := parseActionString(line)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("Action: %s, Type: %s, Name: %s, Size %s\n", action, type_, name, size)

		if action == "CREATE" {
			new_name := strings.Replace(name, "../", "../go-client-monitor/", 1)
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

				conn.Write([]byte("READY")) // send readiness confirmation after processing the line

				// Stream data from the connection to the file
				_, err = io.CopyN(file, conn, bytesToRead)
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
				fmt.Printf("Dir Created %s\n", name)

			default:
				fmt.Printf("No action taken for type %s\n", type_)
			}

			conn.Write([]byte("OK")) // send done confirmation after processing the line
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
