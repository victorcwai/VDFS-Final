func getFileFromClient(fileName string, connection net.Conn,bufferChan chan []byte, stringChan chan string) { //put the file in memory
	stringChan<-"send"
	fileBuffer := make([]byte, BUFFER_SIZE)

	var err error
	// file, err := os.Create(strings.TrimSpace(fileName))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// make a read buffer
	r := bufio.NewReader(connection)
	// // make a write buffer
	// w := bufio.NewWriter(file)

	for err == nil || err != io.EOF {

		// connection.Read(fileBuffer)

		// cleanedFileBuffer := bytes.Trim(fileBuffer, "\x00")

		// _, err = file.WriteAt(cleanedFileBuffer, currentByte)
		// if len(string(fileBuffer)) != len(string(cleanedFileBuffer)) {
		//     break
		// }
		// currentByte += BUFFER_SIZE

		//read a chunk
		n, err := r.Read(fileBuffer)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}
		// // write a chunk
		// if _, err := w.Write(fileBuffer[:n]); err != nil {
		// 	panic(err)
		// }
		bufferChan<-fileBuffer[:n]
		stringChan<-fileName
		// _, err = file.ReadAt(fileBuffer, currentByte)
		// currentByte += BUFFER_SIZE
		// fmt.Println(fileBuffer)
		// connection.Write(fileBuffer)
	}

	// if err = w.Flush(); err != nil {
	// 	panic(err)
	// }
	connection.Close()
	// file.Close()
	return 

}

func WriteFileHandler(connection net.Conn,fileBytes []byte){
	fmt.Printf("%v",(fileBytes))
	connection.Write(fileBytes)
	// var err error
	// w := bufio.NewWriter(connection)
	// if _, err := w.Write(fileBytes); err != nil {
	// 	panic(err)
	// }
	// if err = w.Flush(); err != nil {
	// 	panic(err)
	// }
	connection.Close()
	fmt.Println("Finished sending.")
}
