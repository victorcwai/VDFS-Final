 // volume SERVER //

package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	// "runtime"
	// "strconv"
	"bytes"
	"strings"
	"time"
	"log"
	"os"
	"encoding/gob"
	"github.com/streamrail/concurrent-map"
)

const BUFFERSIZE = 65536
const CHUNKSIZE int64 = 134217728 //2^27

func main() {
    mapping := cmap.New() //{chunk id, chunk bytes}
	fileChunkMap := cmap.New() //{file id, chunks count}
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
		go connectionHandler(connection,mapping,fileChunkMap)

	}
}

func connectionHandler(connection net.Conn, mapping cmap.ConcurrentMap, fileChunkMap cmap.ConcurrentMap) {
	start := time.Now()
	fmt.Println("Connected")

	bufferCommand := make([]byte, 4)
	bufferChunkName := make([]byte, 64)

	connection.Read(bufferCommand)
	command := string(bufferCommand)
	fmt.Println("Command is: " + command)

	//handle reuqest of sending/getting chunk list
	if(command == "setL"){
		setChunkList(connection,fileChunkMap,start)
		return
	}else if(command == "getL"){
		getChunkList(connection,fileChunkMap,start)
		return
	}

	//handle request of sending/getting chunks
	connection.Read(bufferChunkName)
	chunkName := string(bufferChunkName)
	chunkName = strings.Trim(string(chunkName), ":")

	fmt.Println("chunkName is: " + chunkName)
	
	if (command == "send"){
		receiveChunk(chunkName,connection,mapping,start)
		return
	}else if (command == "rece"){
		sendChunk(chunkName,connection,mapping,start)
		return
	}else if (command == "dele") {
		removeChunk(chunkName,mapping)
		connection.Close()
		return
	}else {
		fmt.Println("Bad command received, connectionHandler finished")
		connection.Close()
		return
	}
}

func setChunkList(connection net.Conn, fileChunkMap cmap.ConcurrentMap, start time.Time){
	//then receive chunkList from client and update fileChunkMap
    dec := gob.NewDecoder(connection)
	
    var chunkList []string

    dec.Decode(&chunkList)
    fmt.Printf("chunkList received : %v\n", chunkList);

    fileChunkMap.Set(chunkList[0],chunkList[1]) //chunkList[0] = ID
    connection.Close()
}

func getChunkList(connection net.Conn, fileChunkMap cmap.ConcurrentMap, start time.Time){

	//send chunks location (fileChunkMap[ID]) to client
	var fileName string
	dec := gob.NewDecoder(connection)
	dec.Decode(&fileName)

	fmt.Println("File name is: " + fileName)

	enc := gob.NewEncoder(connection)
	chunkCount,ok := fileChunkMap.Get(fileName)
	if(ok==false){
		fmt.Println("File not found. Request finished.")
		return
	}
	err := enc.Encode(chunkCount)
    if err != nil {
    	log.Fatal("encode error:", err)
    }
}

func receiveChunk(chunkName string, connection net.Conn, mapping cmap.ConcurrentMap, start time.Time){
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

	// for k := range mapping.Iter() {
 //    	fmt.Printf("%s\n", k.)
	// }

	// var m runtime.MemStats  
 //    runtime.ReadMemStats(&m)
 //    fmt.Printf("%d,%d,%d,%d\n", m.HeapSys, m.HeapAlloc, m.HeapIdle, m.HeapReleased)
	// log.Println("Memory Acquired: ", m.Sys)
	// log.Println("Memory Used    : ", m.Alloc)
	connection.Close()
}

func sendChunk(chunkName string, connection net.Conn, mapping cmap.ConcurrentMap, start time.Time){
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
	fmt.Println("Finished sending.")
	connection.Close()
}

func removeChunk(chunkName string, mapping cmap.ConcurrentMap){
	mapping.Remove(chunkName)
	fmt.Println("Chunk", chunkName ,"deleted successfully.")
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
