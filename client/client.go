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
	"bytes"
	"encoding/gob"
)
//64512
const BUFFERSIZE = 65536
const CHUNKSIZE int64 = 134217728 //2^27

func main() {
	//for {
		//read from user
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Please enter 'rece <filename>' or 'send <filename>' to transfer files to the server\n\n")
		inputFromUser, _ := reader.ReadString('\n')
		arrayOfCommands := strings.Split(inputFromUser, " ")
		arrayOfCommands[1] = strings.Replace(arrayOfCommands[1],"\n","",-1)
	
		//connection to master
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

	// fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	// fileName = fillString(fileInfo.Name(), 64)
	fmt.Println("Sending command to master")
	connection.Write([]byte("send"))
	// connection.Write([]byte(fileName))
	// connection.Write([]byte(fileSize))
	enc := gob.NewEncoder(connection) // Will write to network.
	dec := gob.NewDecoder(connection) // Will read from network.
	var volumeServerMap map[int]string
	dec.Decode(&volumeServerMap)
	fmt.Println(volumeServerMap)
	sendBuffer := make([]byte, BUFFERSIZE)
	chunkCount := 0
	for i := fileInfo.Size(); i > 0; i -= CHUNKSIZE{
		chunkCount+=1
	}
	fmt.Println("chunkCount=",chunkCount)
	chunkList := make([]string, chunkCount+1)
	chunkList[0] = fileInfo.Name()
	fmt.Println("Start sending file")
	var sentByte int64
	//connect to node, for each chunk, send to different volume server
	for i:= 0; i < chunkCount; i += 1{
		index := i%len(volumeServerMap) //to prevent index out of bound
		fmt.Println("index:",index)
		vsConnection, err := net.Dial("tcp", volumeServerMap[index])
		
		vsConnection.Write([]byte("send"))
		
		chunkList[i+1] = volumeServerMap[index]
		
		// vsEnc := gob.NewEncoder(vsConnection) // Will write to network.
		// chunkName := fileInfo.Name()+strconv.Itoa(i)
		// hashedChunkName := hash(chunkName)
		// vsEnc.Encode(hashedChunkName) //send file name first
		chunkName := fillString(fileInfo.Name()+strconv.Itoa(i), 64)
		vsConnection.Write([]byte(chunkName))

		fmt.Println("chunkName:",chunkName)
		// fmt.Println("hashedChunkName:",hashedChunkName)

		sentByte=0
		//send until 1 chunk is completed, then move on to next volume server
		for { 
			_, err = file.Read(sendBuffer)
			if err == io.EOF {
				break
			}
			n, err2 := vsConnection.Write(sendBuffer)
			if err2 != nil {
				fmt.Println(err2)
				return
			}
			if(n!=65536){
				fmt.Println("n",n)
			}
			sentByte = sentByte+int64(n)
			if(sentByte == CHUNKSIZE){
				break;
			}
		}
		vsConnection.Close()
	}
	file.Close()
	fmt.Println("Finished sending.")
	//send the chunk list back to master
	enc.Encode(chunkList)
	connection.Close()
	elapsed := time.Since(start)
    fmt.Printf("Sending flie took %s\n", elapsed)
    return
}

func getFileFromServer(fileName string, connection net.Conn, start time.Time) {

    fmt.Println("receive from server")

	connection.Write([]byte("rece"))

	enc := gob.NewEncoder(connection) // Will write to network.
	enc.Encode(fileName)

	elapsed := time.Since(start)
    fmt.Printf("Sending request took %s\n", elapsed)
	
	var chunkList []string
	dec := gob.NewDecoder(connection) // Will write to network.
	dec.Decode(&chunkList)

	fmt.Println("chunkList:",chunkList)
	// bufferFileSize := make([]byte, 10)

	// connection.Read(bufferFileSize)
	// fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)
	file, err := os.Create(fileName+"New")
	if err != nil {
		log.Fatal(err)
	}

	var receivedBytes int64
	
	for i:= 0; i < len(chunkList); i += 1{
		vsConnection, _ := net.Dial("tcp", chunkList[i])
		vsConnection.Write([]byte("rece"))
		// vsEnc := gob.NewEncoder(vsConnection) // Will write to network.
		chunkName := fillString(fileName+strconv.Itoa(i),64)
		vsConnection.Write([]byte(chunkName))
		// vsEnc.Encode(chunkName) //send file name first
		fmt.Println("chunkName:",chunkName)
		receivedBytes=0
		//send until 1 chunk is completed, then move on to next volume server
		for { 
			n,err := io.CopyN(file, vsConnection, BUFFERSIZE)
			receivedBytes += int64(n)
			if err != nil {
				fmt.Println("err", err.Error())
				fmt.Println("err chunkName",chunkName)
				break
			}
			if(CHUNKSIZE<=receivedBytes){
				break
			}
		}
		vsConnection.Close()
		fmt.Println("Done chunk",i)
	}
	fmt.Println("Receiving file complete.")

	file.Close()
	connection.Close()
    fmt.Println("Finished receiving.")
	return

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

// END CLIENT //