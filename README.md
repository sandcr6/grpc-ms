![rx-m LLC][RX-M LLC]


# Microservices


## Developing a gRPC Microservice

gRPC is a modern RPC system that allows interfaces to evolve gracefully, often without impacting existing clients, much
like a REST API. Because gRPC is based on HTTP/2 it is easy to integrate with common web based tools and infrastructure.
Perhaps one of the best things about gRPC is that it uses the powerful but concise Protocol Buffers IDL (interface
description language). IDLs make describing application interfaces a pleasure. Oh, also it doesn't hurt to mention that
gRPC is fast, typically an order of magnitude faster than the equivalent REST API.

In this lab, we will construct three Golang-based programs, a library, a server, and a client. We will use RPC to
communicate between the client and server, where the library provides us our message protocol.


### 1. Setting up a Golang Build Environment

We'll build our microservice in Go. Go is the language Kubernetes and Docker are written in, and though not required,
writing gRPC services in Go is convenient.

We'll use a `golang` container named `grpcapp` as our build and run environment, and mount the container's working
directory to a folder in our home directory:

```
~$ mkdir ~/grpc && cd ~/grpc

~/grpc$ docker container run -it --name grpcapp --hostname build -v /home/ubuntu/grpc:/data -w /data golang:1.16

root@build:/data#
```

The container run command has the following switches:

- `-it` runs the container with an interactive terminal session, so running this command will attach your terminal
- `--name grpcapp` names our container grpcapp
- `--hostname build` names our hostname build
- `-v /home/ubuntu/grpc:/data` switch mounts the `grpc` folder in our home directory to the container's `/data` directory
- `-w /data` overrides Image default working directory (which for this image is '/go')

Test your Go container's installation by invoking the `go`:

```
root@build:/data# go version

go version go1.16.3 linux/amd64

root@build:/data#
```

Great! Go is installed and ready to go (har)!


### 2. Install Probto Buffer Compiler (Protoc)

Our application will be sending Protobuff encoded message back and forth. To do this, we use the Protobuff compiler
known as 'protoc'.


#### Protoc


We install as follows.

```
root@build:/data# wget https://github.com/protocolbuffers/protobuf/releases/download/v3.15.8/protoc-3.15.8-linux-x86_64.zip

root@build:/data# apt update && apt install -y unzip

root@build:/data# unzip protoc-3.15.8-linux-x86_64.zip -d /protoc

root@build:/data# /protoc/bin/protoc --version

libprotoc 3.15.8

root@build:/data#
```

You can learn more about protocol buffers here https://developers.google.com/protocol-buffers/.


### 3. Golang Module Based Project

We will be create the following programs.

* a library - stored as a Go module under "/mygrpc"
* a server application - stored as a a Go module under "/myserver"
* a client application - stored as a Go module under "/myclient"

Create the directory structure in our build container.

```
root@build:/data# mkdir {/mygrpc,/myserver,/myclient}

root@build:/data#
```

Modules provide Golang developers a way to version a group of one or more packages and track dependenices. You can
learn more about Golang Modules here https://blog.golang.org/using-go-modules.


### 4. Create a Protobuf Library

To define a gRPC service we use protocol buffers (PB). PB service definitions include functions that take a message
and return a message. Messages are like structs or records in other languages and can be composed of various other
types. For this simple example we'll create a service definition that allows users to create Open Source Software
Project entries and retrieve them.

On the VM (not in container), create the following IDL file.

```
~/grpc$ cat ossprojects.proto
syntax = "proto3";

option go_package = "github.com/myname/myproject/mygrpc";

service OSSProject {
  rpc ListProjects (ProjectName) returns (ProjectTitles) {}
  rpc CreateProject (Project) returns (ProjectCreateStatus) {}
}

message ProjectName {
  string name = 1;
}
message ProjectTitles {
  repeated string name = 1;
  repeated string custodian = 2;
}
message ProjectCreateStatus {
  int32 status = 1;
}
message Project {
    string name = 1;
    string custodian = 2;
    string description = 3;
    int32 inceptionYear = 4;
}
```

This file not only contains a set of messages ("message ProjectName", "message ...") that we will serialize, it also
defines a service interface ("service OSSProject"). The service definition will be used to help generate our server
(for receiving requests and sending responses) and client code (for sending requests to and receiving responses from).
In other words, protobuf focuses on messages, yet allows RPC systems (like gRPC) to integrate.

Back in the build container, our IDL file is now available. We could have created in the container
(after install an editor), but a better practice to is to mount (and ultimately commit your files).

To create a Golang module, we use "go mod" subcommand.

```
root@build:/data# cd /mygrpc/

root@build:/mygrpc# go mod init github.com/myname/myproject/mygrpc

go: creating new go.mod: module github.com/myname/myproject/mygrpc

root@build:/mygrpc# cat go.mod

module github.com/myname/myproject/mygrpc

go 1.16

root@build:/mygrpc#
```

The file "go.mod" is used by Golang tools to track depedencies. Currently we have no code to speak of, so the only
depedency is Go itself ("go mod" uses Golang, and we happen to be on 1.16).

To generate our code from the IDL we need two additional packages.

* protoc-gen-go
* protoc-gen-go-grpc

THe first, protoc-gen-go, is a plugin so "protoc" can generate Go code. Some languages are core to protobufs, Go is not
one of them.

The second, protoc-gen-go-grpc, is a plugin that enables RPC code creation. In our case, the client and server network
code that can send a protobuf message.

We only need to generate the code once, hence its in its own Go modules. We later, refer to this module from the our
server and client business logic.

Download the needed plugins.

```
root@build:/mygrpc# go get google.golang.org/protobuf/cmd/protoc-gen-go@v1.26.0

go: downloading google.golang.org/protobuf v1.26.0
go get: added google.golang.org/protobuf v1.26.0

root@build:/mygrpc# cat go.mod

module github.com/myname/myproject/mygrpc

go 1.16

require google.golang.org/protobuf v1.26.0 // indirect

root@build:/mygrpc# go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0

go: downloading google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
go get: added google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0

root@build:/mygrpc# cat go.mod

module github.com/myname/myproject/mygrpc

go 1.16

require (
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
)

root@build:/mygrpc#
```

Even though we have not generated code, Go is already tracking modules we downloaed while inside our "mygrpc" module.

We are now ready to generate our serialization (idl message) and server/client (idl system) code.

Keeping things higher level, we supply the IDL file (ossporjects.proto) via `-I/data ossprojects.proto`
(yes that is a space). The remainging options (--go_out, --go-grpc_out) are telling us to generate Go code for the
message and service IDL, the addition options (--go_opt=module and --go-grpc_opt=module)) are doing string search and
replace. That is needed because Golang has changed how code is pacakged over the years.

```
root@build:/mygrpc# /protoc/bin/protoc --go_out=$(pwd) --go_opt=module="github.com/myname/myproject/mygrpc" --go-grpc_out=$(pwd) --go-grpc_opt=module="github.com/myname/myproject/mygrpc" -I/data ossprojects.proto

root@build:/mygrpc# ls -1

go.mod
go.sum
ossprojects.pb.go
ossprojects_grpc.pb.go

root@build:/mygrpc#
```

We see our code has now been generated (take a look if you dare)!

Next we download transistive dependencies. While we manually downloaded the plugins, the generated Go code has its own
(via import statements), which we need to retrieve. Fortunately, "go mod tidy" does just that, it reads the code and
figures out what needs to be downloaded.

```
root@build:/mygrpc# go mod tidy

go: downloading github.com/google/go-cmp v0.5.5
go: downloading golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
go: finding module for package google.golang.org/grpc/codes
go: finding module for package google.golang.org/grpc
go: downloading google.golang.org/grpc v1.37.0
go: finding module for package google.golang.org/grpc/status
go: found google.golang.org/grpc in google.golang.org/grpc v1.37.0
go: found google.golang.org/grpc/codes in google.golang.org/grpc v1.37.0
go: found google.golang.org/grpc/status in google.golang.org/grpc v1.37.0
go: downloading google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
go: downloading github.com/golang/protobuf v1.5.0
go: downloading golang.org/x/net v0.0.0-20190311183353-d8887717615a
go: downloading golang.org/x/sys v0.0.0-20190215142949-d0b11bdaac8a
go: downloading golang.org/x/text v0.3.0

root@build:/mygrpc# go build

root@build:/mygrpc#
```

The rpc methods defined in the service definition specify their request and response types. When a field is marked as
`repeated`, it is allowed to be repeated any number of times (0 or more), much like passing an array or list. The
messages we have defined make use of primitive types, though they need not.  The integer assignments at the end of each
line in the proto file designate the field ids of the messages. This allows the fields to be identified on the wire
without having to pass large field names.

Since we can build the library, we are ready to move to "our" code.


### 5. Server Module

Our service will listen on port 50088 on all interfaces. We use the pb generated RegisterOSSProjectServer to register
our service with the gRPC server. Once running clients can connect and make gRPC calls to either of the two methods.

Back in your VM (not the container), create our server program.

```
~/grpc$ cat server.go
package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "github.com/myname/myproject/mygrpc"
)

const (
	port = ":50088"
)

type server struct{
	pb.UnimplementedOSSProjectServer
}

func (s *server) ListProjects(ctx context.Context, in *pb.ProjectName) (*pb.ProjectTitles, error) {
	log.Printf("Received: %v", in.Name)
	names := []string{in.Name}
	custs := []string{"cncf"}
	return &pb.ProjectTitles{Name: names, Custodian: custs}, nil
}

func (s *server) CreateProject(ctx context.Context, in *pb.Project) (*pb.ProjectCreateStatus, error) {
	log.Printf("Received: %v", in.Name)
	return &pb.ProjectCreateStatus{Status: 0}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterOSSProjectServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
~/grpc$
```

Of note, `pb "github.com/myname/myproject/mygrpc"`, includes our library we created earlier. Notice the IDL "service"
section and the names, do you see them used in our code?

Back in the container.

```
root@build:/mygrpc# cd ../myserver/

root@build:/myserver# ls -1 /data

ossprojects.proto
protoc-3.15.8-linux-x86_64.zip
server.go

root@build:/myserver#
```

As before, we create a Go module, this time for the server code.

```
root@build:/myserver# go mod init github.com/myname/myproject/myserver

go: creating new go.mod: module github.com/myname/myproject/myserver

root@build:/myserver# cat go.mod

module github.com/myname/myproject/myserver

go 1.16

root@build:/myserver# ls -1

go.mod

root@build:/myserver# cp ../data/server.go .

root@build:/myserver#
```

To give a little more depth on the Go tooling, lets try to compile without setting up our server depedencies.

```
root@build:/myserver# go build

server.go:9:2: no required module provides package github.com/myname/myproject/mygrpc; to add it:
	go get github.com/myname/myproject/mygrpc
server.go:8:2: no required module provides package google.golang.org/grpc; to add it:
	go get google.golang.org/grpc

root@build:/myserver#
```

Notice we get two errors, one is our own library(?), and two is for gRPC. The reason our locally developed library appears missing is expected. In Go, it is expected you share your modules (formerly packages) via a distributed version control system (dvcs) (ala GitHub). Since we are locally developing, we skipped that step of uploading our library. Will "go mod" save the day?

```
root@build:/myserver# go mod tidy

go: finding module for package google.golang.org/grpc
go: finding module for package github.com/myname/myproject/mygrpc
go: found google.golang.org/grpc in google.golang.org/grpc v1.37.0
go: downloading github.com/golang/protobuf v1.4.2
go: downloading github.com/google/go-cmp v0.5.0
go: downloading google.golang.org/protobuf v1.25.0
go: finding module for package github.com/myname/myproject/mygrpc
github.com/myname/myproject/myserver imports
	github.com/myname/myproject/mygrpc: cannot find module providing package github.com/myname/myproject/mygrpc: module github.com/myname/myproject/mygrpc: git ls-remote -q origin in /go/pkg/mod/cache/vcs/0d6140616ee4ad4daf607c8d1cc46b1ccb3246200e8150ed418c68510ac6796b: exit status 128:
	fatal: could not read Username for 'https://github.com': terminal prompts disabled
Confirm the import path was entered correctly.
If this is a private repository, see https://golang.org/doc/faq#git_https for additional information.

root@build:/myserver#
```

It did not. While it download the necessary gRPC librarys (modules), it still failed to find our own local library (see error about github?). To allow for local development, we need to tell our server "go.mod" where to find our local library, we do that via "go mod edit".

```
root@build:/myserver# go mod edit -replace=github.com/myname/myproject/mygrpc=../mygrpc

root@build:/myserver# cat go.mod

module github.com/myname/myproject/myserver

go 1.16

replace github.com/myname/myproject/mygrpc => ../mygrpc

root@build:/myserver# go mod tidy

go: finding module for package google.golang.org/grpc
go: found github.com/myname/myproject/mygrpc in github.com/myname/myproject/mygrpc v0.0.0-00010101000000-000000000000
go: found google.golang.org/grpc in google.golang.org/grpc v1.37.0

root@build:/myserver#
```

We are ready to see if we can build our code.

```
root@build:/myserver# go build

root@build:/myserver#
```

Great the server is up and running! Two down, one to go, now lets create our client.


### 6. Client Module


Similar to our server code, we will create client program and install depedencies.

Back in your VM (not the container).

```
~/grpc$ cat client.go

package main

import (
        "context"
        "log"
        "os"
        "time"

        "google.golang.org/grpc"
	pb "github.com/myname/myproject/mygrpc"
)

const (
        port = ":50088"
)

func main() {
        host := os.Args[1]
        req := os.Args[2]
        conn, err := grpc.Dial(host+port, grpc.WithInsecure())
        if err != nil {
                log.Fatalf("did not connect: %v", err)
        }
        defer conn.Close()
        c := pb.NewOSSProjectClient(conn)

        name := "fluentd"
        if len(os.Args) > 2 {
                name = req
        }
        ctx, cancel := context.WithTimeout(context.Background(), time.Second)
        defer cancel()
        r, err := c.ListProjects(ctx, &pb.ProjectName{Name: name})
        if err != nil {
                log.Fatalf("could not get project: %v", err)
        }
        log.Printf("Projects: %v", r)
}

~/grpc$
```

Again, notice the import statements and calls to the server (via c.ListProjects). For the record, we are not using all
the IDL specification in this client (or server) example.

Create the client module.

```
root@build:/myserver# cd ../myclient/

root@build:/myclient# go mod init github.com/myname/myproject/myclient

go: creating new go.mod: module github.com/myname/myproject/myclient

root@build:/myclient#
```

Copy code over.

```
root@build:/myclient# cp ../data/client.go .

root@build:/myclient# ls -1

client.go
go.mod
root@build:/myclient# cat go.mod
module github.com/myname/myproject/myclient

go 1.16

root@build:/myclient#
```

Setup depedencies.

```
root@build:/myclient# go mod edit -replace=github.com/myname/myproject/mygrpc=../mygrpc

root@build:/myclient# go mod tidy

go: finding module for package google.golang.org/grpc
go: found github.com/myname/myproject/mygrpc in github.com/myname/myproject/mygrpc v0.0.0-00010101000000-000000000000
go: found google.golang.org/grpc in google.golang.org/grpc v1.37.0

root@build:/myclient#
```

Confirm we can compile.

```
root@build:/myclient# go build

root@build:/myclient#
```

Excellent, we are ready to use our RPC based application.


### 7. Running Our Code

From the VM, we launch each program, first the server.

```
~/grpc$ docker container exec -w /myserver grpcapp go run server.go
```

If it works, you won't get the prompt back as its not running as a backgroun daemon.

Next, we exercise our client from another terminal.

```
~/grpc$ docker container exec -w /myclient grpcapp go run client.go localhost fluentd

2021/05/07 00:04:45 Projects: name:"fluentd" custodian:"cncf"

~/grpc$
```

Back on the server you should see something similar.

```
2021/05/07 00:04:45 Received: fluentd
```

We tell our gRPC client to connect to the server (running locally within the container) and pass `fluentd` as our
parameter. The request is sent to the server, which the server processes, replies to, and has the client return the
expected output.

We've now successfully coded a complete gRPC server and client!


### 8. Clean Up

As you can imagine, there is more to RPC (messaging and server code), including versioning, conversion/routing,
and more.

Stop the server program.

```
~/grpc$ docker container exec -w /myserver grpcapp go run server.go
2021/05/07 00:04:45 Projects: name:"fluentd" custodian:"cncf"
^c
~/grpc$
```

Back in the build container, we can simply exit.

```
root@build:/myclient# exit

~/grpc$
```

Finally, remove the build container.

```
~/grpc$ docker container rm grpcapp

~/grpc$
```

