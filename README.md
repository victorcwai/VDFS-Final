# VDFS
In-memory distributed file system

Mainly inspiried by seaweedfs, HDFS, memcached and RAMCloud.

The project, VDFS, is a virtualized DFS system where the whole DFS resides inside RAM, i.e. all metadata and data are in-memory. It will provide ultra low-latency access while being able to scale up efficiently, i.e. able to storing very large objects (larger than GBs) by adding more RAMs.

VDFS consist of 2 types of servers. Volume server and Master server.
Each Volume server stores the chunks of different objects. 
Master server provides the list of Volume servers to client. It also provides mapping for objects and the location of their chunks.

Flow of VDFS

Upload:
Client splits the file into 128mb (134217728 bytes) chunks
Client gets a list of volume servers <with id, url> from master
Client sends each chunks to different data nodes
Client sends master a list which contains the location of the chunks

Access/Download:
Client asks master for the file, query with file name
Master sends back the list containing the location of the chunks
Client uses the list to contact the data nodes in the list
Volume servers send the chunks to client
Client read the data sequentially

Implementation

TODO:
Client split to chunks /
Master server /

send: (upload)
Client send chunkName and chunks according to the volume list from master /
Client send list back to master /
Volume receive chunks in map[chunksID]chunks /

receive: (download)
Client ask master for chunk list with file name /
Client get chunk from volumes /
Volume send chunks /
Client put chunk together /

Evaluation and documentation

TODO:
test file correctness
test multiple volume server
test in cloud/ real environment/multi instance
get log and produce stat and graph
make report/slides: draw protocol, specify problems, comparison of different DFS, see others’ work

Future improvements

Cache file mapping and volume servers mapping
Last chunk can be less than 128mb
Concurrent read/write?
