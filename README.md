# VDFS
In-memory distributed file system

Mainly inspiried by seaweedfs, HDFS, memcached and RAMCloud.

The project, VDFS, is a virtualized DFS system where the whole DFS resides inside RAM, i.e. all metadata and data are in-memory. It will provide ultra low-latency access while being able to scale up efficiently, i.e. able to storing very large objects (larger than GBs) by adding more RAMs.

VDFS consist of 2 types of servers. Volume server and Master server.
Each Volume server stores the chunks of different objects. 
Master server provides the list of Volume servers to client. It also provides mapping for objects and the location of their chunks.

## Flow of VDFS

Upload:
  
1. Client splits the file into 128mb (134217728 bytes) chunks
2. Client gets a list of volume servers <with id, url> from master
3. Client sends each chunks to different data nodes
4. Client sends master a list which contains the location of the chunks

Access/Download:

1. Client asks master for the file, query with file name
2. Master sends back the list containing the location of the chunks
3. Client uses the list to contact the data nodes in the list
4. Volume servers send the chunks to client
5. Client read the data sequentially

## Implementation

### TODO:
- [x] Client split to chunks /
- [x] Master server /

send: (upload)
- [x] Client send chunkName and chunks according to the volume list from master /
- [x] Client send list back to master /
- [x] Volume receive chunks in map[chunksID]chunks /

receive: (download)
- [x] Client ask master for chunk list with file name /
- [x] Client get chunk from volumes /
- [x] Volume send chunks /
- [x] Client put chunk together /

## Evaluation and documentation

### TODO:
- [x] test file correctness /
- [x] test multiple volume server /
- [x] test in cloud/ real environment/multi instance /
- [ ] get log and produce stat and graph 
- [ ] make report/slides: draw protocol, specify problems, comparison of different DFS, see others’ work
- [ ] delete -> recycle buffer
- [ ] locking
- [ ] Change send buffer during last send to be not 65536 bytes
- [ ] concurrently write read

## Future improvements

Cache file mapping and volume servers mapping
Last chunk can be less than 128mb
Concurrent read/write?
