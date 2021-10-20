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
