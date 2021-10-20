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
