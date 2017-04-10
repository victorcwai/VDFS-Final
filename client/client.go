 // CLIENT //

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	// "strings"
	"strconv"
	"time"
	"bytes"
	"sync"
	"encoding/gob"
	// "math/rand"
	"hash/fnv"
)

const BUFFERSIZE = 65536
const CHUNKSIZE int64 = 134217728 //2^27
var volumeServerList = []string{"localhost:2001","localhost:2002"}

type receivedBlock struct {
    data []byte
    id  int
}

func main() {	
		//connection to master
		start := time.Now()
		//connection, err := net.Dial("tcp", os.Args[1])
		// if err != nil {
		// 	fmt.Println("There was an error making a connection")
		// }

		if os.Args[1] == "send" {
			SendFileToServer(os.Args[2], start)
		} else if os.Args[1] == "rece" {
			GetFileFromServer(os.Args[2], start)
		} else if os.Args[1] == "dele" {
			DeleteFileInServer(os.Args[2], start)
		} else {
			fmt.Println("Bad Command")
		}
	//}
}

func SendFileToServer(fileName string, start time.Time) {

	fmt.Println("Send to server")
	var err error

	//file to read
	file, err := os.Open(fileName) // For read access.
	fmt.Println(fileName)
	if err != nil {
		// connection.Write([]byte("-1"))
		log.Fatal(err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}

	// fmt.Println("Sending command to master server.")
	// connection.Write([]byte("send"))
	// enc := gob.NewEncoder(connection) // Will write to network.
	// dec := gob.NewDecoder(connection) // Will read from network.
	// var volumeServerList map[int]string
	// dec.Decode(&volumeServerList)
	// fmt.Println(volumeServerList)
	chunkCount := 0

	// file.Close()

	for i := fileInfo.Size(); i > 0; i -= CHUNKSIZE{
		chunkCount+=1
	}
	chunkList := make([]string, chunkCount+1)
	chunkList[0] = fileInfo.Name()

	fmt.Println("chunkCount=",chunkCount)
	fmt.Println("Start sending file")

	//connect to node, for each chunk, send to different volume server
	for i:= 0; i < chunkCount; i += 1{
		//send until 1 chunk is completed, then move on to next chunk and volume server
		chunkName := fileInfo.Name()+strconv.Itoa(i)
		index := hash(chunkName)% uint32(len(volumeServerList))
		
		fmt.Println("chunkName:",chunkName)  
		fmt.Println("index:",index)

		sendBuffer := make([]byte, BUFFERSIZE)
		vsConnection, _ := net.Dial("tcp", volumeServerList[index])
		fmt.Println(volumeServerList[index])
		vsConnection.Write([]byte("send"))
		
		chunkList[i+1] = volumeServerList[index]
		
		vsConnection.Write([]byte(fillString(chunkName, 64)))      
		var sentByte int64
		sentByte=0
		for { 
			n, err := file.Read(sendBuffer)
			if err != nil {
				fmt.Println("err",err,chunkName)
				break
			}

			n2, err2 := vsConnection.Write(sendBuffer[:n])
			if err2 != nil {
				fmt.Println("err2",err2,chunkName)
				break
			}
			if(n2!=65536){
				fmt.Println("n2",n2)
			}
			sentByte = sentByte+int64(n)
			if(sentByte == CHUNKSIZE){
				break;
			}
		}
		fmt.Println("Finished sending",chunkName)
		vsConnection.Close()
	}
	fmt.Println("Finished sending.")
	file.Close()
			
	//send chunkList to a server
	index2 := hash(fileName)%uint32(len(volumeServerList))
	connection, _ := net.Dial("tcp", volumeServerList[index2])
	connection.Write([]byte("setL"))
	enc := gob.NewEncoder(connection)
	enc.Encode(chunkList)
	connection.Close()

	elapsed := time.Since(start)
    fmt.Printf("Sending flie took %s\n", elapsed)
    return
}

func GetFileFromServer(fileName string, start time.Time) {
    fmt.Println("Receive from server")
	
	index := hash(fileName)%uint32(len(volumeServerList))
	connection, _ := net.Dial("tcp", volumeServerList[index])

	connection.Write([]byte("getL"))

	enc := gob.NewEncoder(connection)
	enc.Encode(fileName)
	
	var chunkList []string
	dec := gob.NewDecoder(connection)
	dec.Decode(&chunkList)
	if(chunkList==nil){
		fmt.Println("File not found.")
		return
	}
	connection.Close()

	fmt.Println("chunkList:",chunkList)

	c := make(chan *receivedBlock)
	var wg sync.WaitGroup
	wg.Add(1)
	go blockAssembler(c,len(chunkList),fileName,&wg,start)
	
	maxGoroutines := 3
    guard := make(chan struct{}, maxGoroutines)

	for i:= len(chunkList)-1; i >= 0; i -= 1{
		wg.Add(1)
		go func(i int,c chan *receivedBlock){
			defer wg.Done()
			guard <- struct{}{} // would block if guard channel is already filled

			vsConnection, _ := net.Dial("tcp", chunkList[i])
			vsConnection.Write([]byte("rece"))
			var bufferFile bytes.Buffer
		    writer := bufio.NewWriter(&bufferFile)
			chunkName := fileName+strconv.Itoa(i)
			vsConnection.Write([]byte(fillString(chunkName,64)))
			fmt.Println("chunkName:",chunkName)
			var receivedBytes int64
			receivedBytes=0

			//receive until 1 chunk is completed, then move on to next volume server
			for { 
				n,err := io.CopyN(writer, vsConnection, BUFFERSIZE)
				receivedBytes += int64(n)
				if err != nil {
					fmt.Println("err", err.Error(), chunkName)
					break
				}
				if(CHUNKSIZE<=receivedBytes){
					break
				}
			}
			vsConnection.Close()
			blk := receivedBlock{bufferFile.Bytes(), i}
			fmt.Println("Sending to chan")
			c <- &blk
			fmt.Println("Done chunk",i)
			<-guard
		}(i,c)
	}
	wg.Wait()
	fmt.Println("Receiving file complete.")

	// file.Close()
    fmt.Println("Finished receiving.")
	elapsed := time.Since(start)
    fmt.Printf("Receiving flie in disk took %s\n", elapsed)
	return

}

func DeleteFileInServer(fileName string, start time.Time) {	    
	fmt.Println("Delete from server")
	
	index := hash(fileName)%uint32(len(volumeServerList))
	connection, _ := net.Dial("tcp", volumeServerList[index])

	connection.Write([]byte("getL"))

	// enc := gob.NewEncoder(connection) 
	// enc.Encode(fileName)
	
	var chunkList []string
	dec := gob.NewDecoder(connection)
	dec.Decode(&chunkList)
	connection.Close()

	fmt.Println("chunkList:",chunkList)
	for i:= 0; i < len(chunkList); i += 1{
		vsConnection, _ := net.Dial("tcp", chunkList[i])
		vsConnection.Write([]byte("dele"))
		chunkName := fillString(fileName+strconv.Itoa(i),64)
		vsConnection.Write([]byte(chunkName))
		fmt.Println("chunkName:",chunkName)
		vsConnection.Close()
		fmt.Println("Done chunk",i)
	}
	fmt.Println("Deleting file complete.")

    fmt.Println("Finished deleting.")
    elapsed := time.Since(start)
    fmt.Printf("Sending request for deleting flie took %s\n", elapsed)

	return

}

func blockAssembler(c chan *receivedBlock, chunkNum int, fileName string, wg *sync.WaitGroup,start time.Time){
	defer wg.Done()
	receivedBlockList := make([]*receivedBlock, chunkNum)
	
	dwCh := make(chan *receivedBlock)
	
	//disk writer
	defer wg.Add(1)
	go func(c2 chan *receivedBlock,cN int){
		defer wg.Done()
		order := 0
		file, err := os.Create("new"+fileName)
		if err != nil {
			log.Fatal(err)
		}

		dwReceivedBlockList := make([]*receivedBlock, cN)
		for i:= 0; i < cN; i += 1{
			blk := <- c2
			dwReceivedBlockList[blk.id] = blk
			if(dwReceivedBlockList[order] != nil) {
				fmt.Println("DW writing chunk no.", order, dwReceivedBlockList[order].id)
				n, err := file.Write(dwReceivedBlockList[order].data)
				if(err!=nil){
					fmt.Println("DW err, n", err.Error(), n)
				}
				order+=1
			}
		}

		//write remaining blocks after every block is received
		for ; order < cN; order += 1{		
			n2, err2 := file.Write(dwReceivedBlockList[order].data)
			fmt.Println("DW writing chunk no.", order, dwReceivedBlockList[order].id)
			if(err2!=nil){
				fmt.Println("DW err2, n2", err2.Error(), n2)
			}
		}
		file.Close()
	}(dwCh,chunkNum)
	
	for i:= 0; i < chunkNum; i += 1{
		blk := <-c
		dwCh <- blk
		receivedBlockList[blk.id] = blk
	}

	fmt.Println("Block Assembler done.")
    elapsed := time.Since(start)
    fmt.Printf("Receiving file in memory took %s\n", elapsed)
}

func hash(s string) uint32 {
        h := fnv.New32a()
        h.Write([]byte(s))
        return h.Sum32()
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