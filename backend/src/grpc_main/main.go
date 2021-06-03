package main;

import (
	"context"
	"net"
	"fmt"
	"time"
	"log"

	"google.golang.org/grpc"
	pb "github.com/htchan/WebHistory/src/protobuf"
)

const (
	port = ":9105"
	database_location = "./database/websites.db"
)

type Server struct {
	pb.UnimplementedWebHistoryServer
}

func (s *Server) Add(ctx context.Context, in *pb.UrlRequest) (*pb.MessageResponse, error) {
	web := Website{Url: in.Url, AccessTime: time.Now().Unix()}
	web.Update()
	web.Save()
	return &pb.MessageResponse{Message: "website <" + in.Url + "> add success"}, nil
}

func (s *Server) List(_ *pb.Empty, stream pb.WebHistory_ListServer) error {
	fmt.Println(Urls())
	for _, url := range Urls() {
		if err := stream.Send(GetWeb(url).Response()); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Refresh(ctx context.Context, in *pb.UrlRequest) (*pb.MessageResponse, error) {
	web := GetWeb(in.Url)
	web.AccessTime = time.Now().Unix()
	fmt.Println("web")
	web.Save()
	return &pb.MessageResponse{Message: "success"}, nil
}

func regularUpdate() {
	for range time.Tick(time.Hour * 23) {
		for _, url := range Urls() {
			web := GetWeb(url)
			web.Update()
			web.Save()
		}
	}
}

func main() {
	openDatabase("./database/websites.db")
	go regularUpdate()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterWebHistoryServer(s, &Server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}