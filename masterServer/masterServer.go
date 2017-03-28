 // master SERVER //

package main

import (
	"fmt"
	"net"	
	"time"
	"encoding/gob"
	"log"
	"github.com/streamrail/concurrent-map"
)

const BUFFERSIZE = 65536
const CHUNKSIZE = 134217728 //2^27

func main() {
    volumeServerMap := make(map[int]string) //{volume server id, url}
    fileChunkMap := cmap.New() //{file id, {chunks url}}

    volumeServerMap[0] = "localhost:2001"
	fmt.Println("Master server start listening")

	server, error := net.Listen("tcp", ":2000")
	if error != nil {
		fmt.Println("There was an error starting the server" + error.Error())
		return
	}

	//infinate loop
	for {

		connection, error := server.Accept()
		if error != nil {
			fmt.Println("There was am error with the connection" + error.Error())
			return
		}
		//handle the connection, on it's own thread, per connection
		go ConnectionHandler(connection,volumeServerMap,fileChunkMap)

	}
}

func ConnectionHandler(connection net.Conn, volumeServerMap map[int]string,fileChunkMap cmap.ConcurrentMap) {
	fmt.Println("Connected")
	bufferCommand := make([]byte, 4)
	connection.Read(bufferCommand)
	command := string(bufferCommand)
	fmt.Println("Command is: " + command)

	if(command=="send"){
		//send volumeServerMapping to client
		//then receive chunkList from client and update chunkMap
		
		enc := gob.NewEncoder(connection) // Will write to network.
	    dec := gob.NewDecoder(connection) // Will read from network.
		
		start := time.Now()
		err := enc.Encode(volumeServerMap)
	    if err != nil {
	    	log.Fatal("encode error:", err)
	    }	

	    var chunkList []string

	    dec.Decode(&chunkList)
	    fmt.Printf("chunkList received : %v\n", chunkList);

	    fileChunkMap.Set(chunkList[0],chunkList[1:]) //chunkList[0] = ID
	    connection.Close()
	    
	    fmt.Println(volumeServerMap)
		fmt.Println(fileChunkMap)

		elapsed := time.Since(start)
		fmt.Printf("Request took %s\n", elapsed)
	}else if(command =="rece"){
		//send chunks location (fileChunkMap[ID]) to client
		start := time.Now()
		var fileName string
		dec := gob.NewDecoder(connection)
		dec.Decode(&fileName)

		fmt.Println("File name is: " + fileName)

		enc := gob.NewEncoder(connection) // Will write to network.
		chunkURLs,ok := fileChunkMap.Get(fileName)
		if(ok==false){
			fmt.Println("File not found. Request finished.")
			return
		}
		err := enc.Encode(chunkURLs)
	    if err != nil {
	    	log.Fatal("encode error:", err)
	    }
		elapsed := time.Since(start)
		fmt.Printf("Request took %s\n", elapsed)
	}else if(command == "dele"){
		start := time.Now()
		var fileName string
		dec := gob.NewDecoder(connection)
		dec.Decode(&fileName)

		fmt.Println("File name is: " + fileName)

		enc := gob.NewEncoder(connection) 
		chunkURLs,_ := fileChunkMap.Get(fileName)
		err := enc.Encode(chunkURLs)
	    if err != nil {
	    	log.Fatal("encode error:", err)
	    }

	    fileChunkMap.Remove(fileName)
		fmt.Println("File", fileName ,"deleted successfully.")
		
		elapsed := time.Since(start)
		fmt.Printf("Request took %s\n", elapsed)
	}

    fmt.Println("Request finished.")

}


// func fillString(returnString string, toLength int) string {
// 	byteBuffer := bytes.NewBufferString(returnString)
// 	lengthString := len(returnString)
// 	count := toLength-lengthString
// 	for {
// 		if count > 0 {
// 			byteBuffer.WriteString(":")
// 			count -= 1
// 			continue
// 		}
// 		break
// 	}
// 	return byteBuffer.String()
// }

// // END master SERVER //