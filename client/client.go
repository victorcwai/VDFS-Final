 // CLIENT //

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"strconv"
	"time"
)

const BUFFERSIZE = 64512

func main() {
	//for {
		//read from user
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Please enter 'rece <filename>' or 'send <filename>' to transfer files to the server\n\n")
		inputFromUser, _ := reader.ReadString('\n')
		arrayOfCommands := strings.Split(inputFromUser, " ")
		arrayOfCommands[1] = strings.Replace(arrayOfCommands[1],"\n","",-1)
	
		//get port and ip address to dial
		connection, err := net.Dial("tcp", "localhost:2000")
		if err != nil {
			fmt.Println("There was an error making a connection")
		}
		start := time.Now()

		if arrayOfCommands[0] == "rece" {
			getFileFromServer(arrayOfCommands[1], connection, start)

		} else if arrayOfCommands[0] == "send" {
			sendFileToServer(arrayOfCommands[1], connection, start)
		} else {
			fmt.Println("Bad Command")
		}
	//}
}

func sendFileToServer(fileName string, connection net.Conn, start time.Time) {

	fmt.Println("Send to server")
	var err error

	//file to read
	file, err := os.Open(fileName) // For read access.
	fmt.Println(fileName)
	if err != nil {
		connection.Write([]byte("-1"))
		log.Fatal(err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}

	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	fileName = fillString(fileInfo.Name(), 64)
	fmt.Println("Sending command, filename and filesize")
	connection.Write([]byte("send"))
	connection.Write([]byte(fileName))
	connection.Write([]byte(fileSize))
	sendBuffer := make([]byte, BUFFERSIZE)
	fmt.Println("Start sending file")

	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		connection.Write(sendBuffer)
	}

	file.Close()
	connection.Close()
	fmt.Println("Finished sending.")
	elapsed := time.Since(start)
    fmt.Printf("Sending flie took %s\n", elapsed)
}

func getFileFromServer(fileName string, connection net.Conn, start time.Time) {

    fmt.Println("receive from server")

	connection.Write([]byte("rece"))
	connection.Write([]byte(fillString(fileName,64)))
	elapsed := time.Since(start)
    fmt.Printf("Sending request took %s\n", elapsed)

	bufferFileSize := make([]byte, 10)

	connection.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)


	var err error

	file, err := os.Create(fileName+"New")
	if err != nil {
		log.Fatal(err)
	}

	var receivedBytes int64
	
	for {
		if (fileSize - receivedBytes) < BUFFERSIZE {
			io.CopyN(file, connection, (fileSize - receivedBytes))
			connection.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
			break
		}
		io.CopyN(file, connection, BUFFERSIZE)
		receivedBytes += BUFFERSIZE
	}
	fmt.Println("Receiving file complete.")

	file.Close()
	connection.Close()
    fmt.Println("Finished receiving.")
	return

}

func fillString(retunString string, toLength int) string {
	for {
		lengtString := len(retunString)
		if lengtString < toLength {
			retunString = retunString + ":"
			continue
		}
		break
	}
	return retunString
}

// END CLIENT //