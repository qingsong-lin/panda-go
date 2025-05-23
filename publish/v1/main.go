package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pubsub "panda-go/publish/v1/pb"
)

// PubSubServer 实现了 PubSub 服务
type PubSubServer struct {
	//*pubsub.UnimplementedPubSubServer
	subscribers map[chan<- *pubsub.Message]struct{}
}

// NewPubSubServer 创建一个新的 PubSub 服务器
func NewPubSubServer() *PubSubServer {
	return &PubSubServer{
		subscribers: make(map[chan<- *pubsub.Message]struct{}),
	}
}

// Subscribe 实现 Subscribe RPC
func (s *PubSubServer) Subscribe(req *pubsub.Message, stream pubsub.PubSub_SubscribeServer) error {
	ch := make(chan *pubsub.Message, 1)
	s.subscribers[ch] = struct{}{}

	defer func() {
		close(ch)
		delete(s.subscribers, ch)
	}()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return status.Error(codes.Unavailable, "subscriber closed")
			}

			if err := stream.Send(msg); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		default:

		}
	}
}

// Publish 实现 Publish RPC
func (s *PubSubServer) Publish(ctx context.Context, req *pubsub.Message) (*pubsub.Message, error) {
	for ch := range s.subscribers {
		for {
			select {
			case ch <- req:
				println("ok write")
			default:
				//println("Drop the message if the channel is full") //
			}
		}
	}

	return req, nil
}

func main() {
	server := grpc.NewServer()
	pubsub.RegisterPubSubServer(server, NewPubSubServer())

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	fmt.Println("Server is listening on :50051")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
