# FileTransfer
transfer file|folder with tcp ，implemented by go

# Client
    Client is the sender in file transfering . All the function packaged in 
SendResourceTo(host,resPath string ,isDir bool). Param "host" is the address 
of server,like "127.0.0.1:7010". Param "resPath" is the path of resource which
will be uploaded. The "isDir",identity the resource which is file(false) or 
folder(true) .

# Server
    Server is the receiver in file transfering. All the function packaged in 
ReceiveResource(conn net.Conn, downloadPath string).Param "conn" is the connection
object by listener.Accept() from client.Param "downloadPath" is the path where to 
storage resource from client.

# Property

> 1. file or folder transfering
> 2. display transmission speed and progress
> 3. breakpoint continual transfer

