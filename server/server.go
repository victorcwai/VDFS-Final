 // SERVER //

package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"runtime"
	"strconv"
	"bytes"
	"strings"
	"time"
)

const BUFFERSIZE = 64512

func main() {
    //pool := make([][]byte, 100)
    mapping := make(map[string][]byte)
    // dataChan := make(chan []byte)
    // stringChan := make(chan string)
    // bufCount := 0
    
	fmt.Println("Server start listening")

	server, error := net.Listen("tcp", ":2000")
	if error != nil {
		fmt.Println("There was an error starting the server" + error.Error())
		return
	}
	// var received string

	//infinate loop
	for {

		connection, error := server.Accept()
		if error != nil {
			fmt.Println("There was am error with the connection" + error.Error())
			return
		}
		//handle the connection, on it's own thread, per connection
		go ConnectionHandler(connection,mapping)

	}
}

func ConnectionHandler(connection net.Conn, mapping map[string][]byte) {
	fmt.Println("Connected")
	
		bufferCommand := make([]byte, 4)
		bufferFileName := make([]byte, 64)

		// _, error := connection.Read(buffer)
		// if error != nil {
		// 	fmt.Println("There is an error reading from connection", error.Error())
		// 	stringChan<-"failed"
		// 	return 
		// }

		connection.Read(bufferCommand)
		command := string(bufferCommand)

		connection.Read(bufferFileName)
		fileName := strings.Trim(string(bufferFileName), ":")

		fmt.Println("Command is: " + command)
		fmt.Println("File name is: " + fileName)
		// fmt.Println("File name is: ",[]byte(fileName))
		
		//loop until disconnect?
		start := time.Now()
		if (string(command) == "send"){
			bufferFileSize := make([]byte, 10)
			connection.Read(bufferFileSize)
			fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)
			
			fmt.Println("File size is: " + string(bufferFileSize))
			//make a buffer to hold data		
		    var bufferFile bytes.Buffer
		    writer := bufio.NewWriter(&bufferFile)

			var receivedBytes int64
		    
			for {
				if (fileSize - receivedBytes) < BUFFERSIZE { //if there is no need for next read,
					io.CopyN(writer, connection, (fileSize - receivedBytes))
					connection.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize)) //to clear remaining bytes in the buffer
					break
				}
				io.CopyN(writer, connection, BUFFERSIZE)
				receivedBytes += BUFFERSIZE
			}
			mapping[fileName] = bufferFile.Bytes()

			elapsed := time.Since(start)
		    fmt.Printf("Storing file took %s\n", elapsed)

			for k, _ := range mapping {
	        	fmt.Printf("%s\n", k)
	    	}
			// fmt.Println(string(pool[bufCount]))
			var m runtime.MemStats  
		    runtime.ReadMemStats(&m)
		    fmt.Printf("%d,%d,%d,%d,%d\n", len(mapping[fileName]), m.HeapSys, m.HeapAlloc, m.HeapIdle, m.HeapReleased)

			// stringChan<-"send"
			// dataChan<-bufferFile.Bytes()
			// stringChan<-fileName
			// connection.Close()

		}else if (string(command) == "rece"){
			//fmt.Println("mapping[fileName] return: %v",mapping[fileName])
			
			bufferFileSize := fillString(strconv.FormatInt(int64(len(mapping[fileName])), 10), 10)
			elapsed := time.Since(start)
		    fmt.Printf("File access took %s\n", elapsed)
			connection.Write([]byte(bufferFileSize))
			fmt.Println("File size is: " + string(bufferFileSize))
			fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

			reader := bytes.NewReader(mapping[fileName])
			var sentBytes int64
			for {
				if (fileSize - sentBytes) < BUFFERSIZE { //if there is no need for next read,
					io.CopyN(connection, reader, (fileSize - sentBytes))
					reader.Read(make([]byte, (sentBytes+BUFFERSIZE)-fileSize)) //to clear remaining bytes in the buffer
					break
				}
				io.CopyN(connection, reader, BUFFERSIZE)
				sentBytes += BUFFERSIZE
			}

			//connection.Write(fileBytes)

			// connection.Close()
			fmt.Println("Finished sending.")

		}else{
			fmt.Println("Bad command received, ConnectionHandler finished")
			return;
		}

	    fmt.Println("Request finished.")
	// }
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

// END SERVER //
