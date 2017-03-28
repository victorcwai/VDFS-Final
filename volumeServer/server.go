 // volume SERVER //

package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"runtime"
	// "strconv"
	"bytes"
	"strings"
	"time"
	"log"
	// "encoding/gob"
	"github.com/streamrail/concurrent-map"
)

const BUFFERSIZE = 65536
const CHUNKSIZE int64 = 134217728 //2^27

func main() {
    //pool := make([][]byte, 100)
    mapping := cmap.New()
    // dataChan := make(chan []byte)
    // stringChan := make(chan string)
    // bufCount := 0

	fmt.Println("Volume server start listening")

	server, error := net.Listen("tcp", ":2001")
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

func ConnectionHandler(connection net.Conn, mapping cmap.ConcurrentMap) {
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

	// dec := gob.NewDecoder(connection) // Will read from network.
	// var fileName string
	// dec.Decode(&fileName)
	connection.Read(bufferFileName)
	fileName := string(bufferFileName)
	fileName = strings.Trim(string(fileName), ":")
	// connection.Read(bufferFileName)
	// fileName := strings.Trim(string(bufferFileName), ":")

	fmt.Println("Command is: " + command)
	fmt.Println("fileName is: " + fileName)
	// fmt.Println("File name is: ",[]byte(fileName))
	
	//loop until disconnect?
	start := time.Now()
	if (command == "send"){
		// bufferFileSize := make([]byte, 10)
		// connection.Read(bufferFileSize)
		
		// fmt.Println("File size is: " + string(bufferFileSize))
		//make a buffer to hold data		
	    var bufferFile bytes.Buffer
	    writer := bufio.NewWriter(&bufferFile)

		var receivedBytes int64
	    receivedBytes=0
		for {
			// if (fileSize - receivedBytes) < BUFFERSIZE { //if there is no need for next read,
			// 	io.CopyN(writer, connection, (fileSize - receivedBytes))
			// 	connection.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize)) //to clear remaining bytes in the buffer
			// 	break
			// }

			if(CHUNKSIZE<=receivedBytes){
				break
			}
			n,err := io.CopyN(writer, connection, BUFFERSIZE)
			receivedBytes += n
			// if(n!=65536){
			// 	fmt.Println("n",n)
			// 	fmt.Println("receivedBytes",receivedBytes)
			// 	fmt.Println("err fileName",fileName)
			// }
			if err != nil {
				fmt.Println("err", err.Error())
				fmt.Println("err fileName",fileName)
				break
			}
			// log.Println("READ: ",receivedBytes)
		}
		mapping.Set(fileName,bufferFile.Bytes())
		//fmt.Println("len(mapping[fileName])",len(mapping.Get(fileName)))
		elapsed := time.Since(start)
	    fmt.Printf("Storing file took %s\n", elapsed)

		for k := range mapping.Iter() {
        	fmt.Printf("%s\n", k)
    	}
		// fmt.Println(string(pool[bufCount]))
		// runtime.GC()
		var m runtime.MemStats  
	    runtime.ReadMemStats(&m)
	    fmt.Printf("%d,%d,%d,%d\n", m.HeapSys, m.HeapAlloc, m.HeapIdle, m.HeapReleased)
		log.Println("Memory Acquired: ", m.Sys)
		log.Println("Memory Used    : ", m.Alloc)

		// stringChan<-"send"
		// dataChan<-bufferFile.Bytes()
		// stringChan<-fileName
		connection.Close()

	}else if (command == "rece"){
		//fmt.Println("mapping[fileName] return: %v",mapping[fileName])
		
		// bufferFileSize := fillString(strconv.FormatInt(int64(len(mapping[fileName])), 10), 10)
		elapsed := time.Since(start)
	    fmt.Printf("File access took %s\n", elapsed)
		// connection.Write([]byte(bufferFileSize))
		// fmt.Println("File size is: " + string(bufferFileSize))
		// fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	    fileData,_ := mapping.Get(fileName)
	    fileBytes := fileData.([]byte)
		reader := bytes.NewReader(fileBytes)
		fileSize := len(fileBytes)
		sendBuffer := make([]byte,BUFFERSIZE)

		var sentBytes int64
		for {

			// if (len(mapping[fileName]) - sentBytes) < BUFFERSIZE { //if there is no need for next read,
			// 	reader.Read(sendBuffer)
			// 	connection.Write(sendBuffer)
			// 	break
			// }
			if (int64(fileSize) - sentBytes <= 0){
				break
			}
			n,err := reader.Read(sendBuffer)			
			if err == io.EOF {
				break
			}

			_,err2 := connection.Write(sendBuffer[:n])
			if err2 != nil {
				fmt.Println("err2", err2.Error())
				fmt.Println("err fileName",fileName)
				break
			}
			sentBytes += int64(n)
		}

		//connection.Write(fileBytes)

		connection.Close()
		fmt.Println("Finished sending.")

	}else if (command == "dele") {
		mapping.Remove(fileName)
		fmt.Println("Chunk", fileName ,"deleted successfully.")
	}else {
		fmt.Println("Bad command received, ConnectionHandler finished")
		return;
	}

    fmt.Println("Request finished.")
// }
}


func fillString(returnString string, toLength int) string {
	byteBuffer := bytes.NewBufferString(returnString)
	lengthString := len(returnString)
	count := toLength-lengthString
	for {
		if count > 0 {
			byteBuffer.WriteString(":")
			count -= 1
			continue
		}
		break
	}
	return byteBuffer.String()
}

// END volume SERVER //