package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	pubsub "panda-go/publish/v2/pb"
)

// PubSubServer 实现了 PubSub 服务
type PubSubServer struct {
	mu           sync.Mutex
	subscribers  map[string]map[chan<- *pubsub.Message]struct{} // 订阅者列表
	messageQueue chan *pubsub.Message                           // 消息队列
}

// NewPubSubServer 创建一个新的 PubSub 服务器
func NewPubSubServer() *PubSubServer {
	return &PubSubServer{
		subscribers:  make(map[string]map[chan<- *pubsub.Message]struct{}),
		messageQueue: make(chan *pubsub.Message, 100), // 设置合适的缓冲区大小
	}
}

// Subscribe 实现 Subscribe RPC
func (s *PubSubServer) Subscribe(req *pubsub.Subscription, stream pubsub.PubSub_SubscribeServer) error {
	ch := make(chan *pubsub.Message, 1)
	s.addSubscriber(req.ClientId, req.Topic, ch)

	defer func() {
		s.removeSubscriber(req.ClientId, req.Topic, ch)
		close(ch)
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
		}
	}
}

// Publish 实现 Publish RPC
func (s *PubSubServer) Publish(ctx context.Context, req *pubsub.Message) (*emptypb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if subscribers, ok := s.subscribers[req.Topic]; ok {
		for ch := range subscribers {
			select {
			case ch <- req:
			default:
				// Drop the message if the channel is full
			}
		}
	}

	return &emptypb.Empty{}, nil
}

// addSubscriber 添加订阅者
func (s *PubSubServer) addSubscriber(clientID, topic string, ch chan<- *pubsub.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.subscribers[topic]; !ok {
		s.subscribers[topic] = make(map[chan<- *pubsub.Message]struct{})
	}

	s.subscribers[topic][ch] = struct{}{}
	fmt.Printf("Client %s subscribed to topic %s\n", clientID, topic)
}

// removeSubscriber 移除订阅者
func (s *PubSubServer) removeSubscriber(clientID, topic string, ch chan<- *pubsub.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if subscribers, ok := s.subscribers[topic]; ok {
		delete(subscribers, ch)
		close(ch)
		fmt.Printf("Client %s unsubscribed from topic %s\n", clientID, topic)
	}
}

// broadcastMessages 定期广播消息
func (s *PubSubServer) broadcastMessages() {
	for {
		select {
		case msg := <-s.messageQueue:
			s.Publish(context.Background(), msg)
		}
	}
}

func main() {
	server := grpc.NewServer()
	pubsub.RegisterPubSubServer(server, NewPubSubServer())

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	fmt.Println("Server is listening on :50051")

	// 启动消息广播协程
	go NewPubSubServer().broadcastMessages()

	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
