package main

import (
    "context"
    "log"
    "net/http"
    "fmt"
    "io"
    pb "connector/gen" // import path to your generated files

    //"google.golang.org/grpc"
)

// server implements ImageServiceServer (from imageservice_grpc.pb.go)
type server struct {
    pb.UnimplementedImageServiceServer
}

func (s *server) Identify(ctx context.Context, req *pb.IdentifyRequest) (*pb.IdentifyResponse, error) {
    log.Printf("Identify called with base_image url: %s", req.BaseImage.Url)

    // Just returning a dummy embedding for demonstration

    //request := &pb.IdentifyRequest{
      //BaseImage: &pb.Image{Url: "hello"},
    //};

    return nil,nil
}

func similarity(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	io.WriteString(w, "Empty response\n")
}

func main() {
  http.HandleFunc("/similarity", similarity)
	err := http.ListenAndServe(":80", nil)

  if err != nil {
    fmt.Println(err)
  }

  return;
}
