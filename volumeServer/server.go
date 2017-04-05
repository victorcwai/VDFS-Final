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
	"os"
	// "encoding/gob"
	"github.com/streamrail/concurrent-map"
)

const BUFFERSIZE = 65536
const CHUNKSIZE int64 = 134217728 //2^27

func main() {
    mapping := cmap.New()

	server, error := net.Listen("tcp", os.Args[1])
	if error != nil {
		fmt.Println("There was an error starting the server" + error.Error())
		return
	}

	fmt.Println("Volume server start listening")

	//infinate loop
	for {

		connection, error := server.Accept()
		if error != nil {
			fmt.Println("There was am error with the connection" + error.Error())
			return
		}
		//one goroutine per connection
		go ConnectionHandler(connection,mapping)

	}
}

func ConnectionHandler(connection net.Conn, mapping cmap.ConcurrentMap) {
	fmt.Println("Connected")

	bufferCommand := make([]byte, 4)
	bufferChunkName := make([]byte, 64)

	connection.Read(bufferCommand)
	command := string(bufferCommand)

	connection.Read(bufferChunkName)
	chunkName := string(bufferChunkName)
	chunkName = strings.Trim(string(chunkName), ":")

	fmt.Println("Command is: " + command)
	fmt.Println("chunkName is: " + chunkName)
	
	start := time.Now()
	if (command == "send"){
		//make a buffer to hold data		
	    var bufferFile bytes.Buffer
	    writer := bufio.NewWriter(&bufferFile)

		var receivedBytes int64
	    receivedBytes=0
		for {

			if(CHUNKSIZE<=receivedBytes){
				break
			}
			n,err := io.CopyN(writer, connection, BUFFERSIZE)
			receivedBytes += n
			if err != nil {
				fmt.Println("err", err.Error(), chunkName)
				break
			}
		}
		mapping.Set(chunkName,bufferFile.Bytes())
		elapsed := time.Since(start)
	    fmt.Printf("Storing file took %s\n", elapsed)

		for k := range mapping.Iter() {
        	fmt.Printf("%s\n", k.Key)
    	}
		var m runtime.MemStats  
	    runtime.ReadMemStats(&m)
	    fmt.Printf("%d,%d,%d,%d\n", m.HeapSys, m.HeapAlloc, m.HeapIdle, m.HeapReleased)
		log.Println("Memory Acquired: ", m.Sys)
		log.Println("Memory Used    : ", m.Alloc)

		connection.Close()

	}else if (command == "rece"){
		elapsed := time.Since(start)
	    fmt.Printf("File access took %s\n", elapsed)
	    fileData,_ := mapping.Get(chunkName)
	    fileBytes := fileData.([]byte)
		reader := bytes.NewReader(fileBytes)
		fileSize := len(fileBytes)
		sendBuffer := make([]byte,BUFFERSIZE)

		var sentBytes int64
		for {
			if (int64(fileSize) - sentBytes <= 0){
				break
			}
			n,err := reader.Read(sendBuffer)			
			if err == io.EOF {
				break
			}

			_,err2 := connection.Write(sendBuffer[:n])
			if err2 != nil {
				fmt.Println("err2", err2.Error(), chunkName)
				break
			}
			sentBytes += int64(n)
		}

		connection.Close()
		fmt.Println("Finished sending.")

	}else if (command == "dele") {
		mapping.Remove(chunkName)
		fmt.Println("Chunk", chunkName ,"deleted successfully.")
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