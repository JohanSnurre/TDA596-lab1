## Start up
##### server:
The file containing server is run with *main.exe*. User just have to run it and specify the server port as an argument. Then the server is running. To do

 - the **get** request - the url format for getting is: 
 *address/<requested_file.extension>*
 - the **post** request - the url format for posting is just:
 *address* 
 To upload the file, the header has to be *multipart/form-data* and the body is should be a form when the key for the file is *file* and the value is wanted file for the upload. If successful, the server return 201.

##### proxy:
The file containing proxy is in *proxy/proxy.exe*. User can run it with arguments, where first argument is the port for the proxy and the second argument is the address for the server. 

## Cloud
We are running one EC2 instance for the server and one for proxy. The server has opened port 12345 to be exposed on and the proxy has 12346 port opened. We just compiled the code into .exe files and run them. The addresses then are:

 - the server:

http://ec2-52-54-103-84.compute-1.amazonaws.com:12345/<requested_file.extension> 

 - the proxy
https://ec2-3-93-43-235.compute-1.amazonaws.com:12346/<requested_file.extension>