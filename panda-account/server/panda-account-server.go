package server

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"os"
	"panda-go/common/log"
	"panda-go/common/utils"
	"panda-go/proto/pb"
)

type PandaAccountServer struct {
}

func (h *PandaAccountServer) SayHello2DemoClient(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{Reply: "i am panda panda-go client"}, nil
}

func NewPandaAccountServer() *PandaAccountServer {
	return &PandaAccountServer{}
}

func (h *PandaAccountServer) UploadFile(ss pb.PandaAccount_UploadFileServer) error {
	return status.Errorf(codes.Unimplemented, "method UploadFile not implemented")
}

func (h *PandaAccountServer) DownloadFile(req *pb.DownloadFileRequest, ss pb.PandaAccount_DownloadFileServer) error {
	if err := utils.WriteStringToFile("lgx.txt", "{\"sdfasdfasdfasdfasdfasdfasdfasdfasd\":\"dfkgl;dajlkfjsadlkfjklsadnfkjasdhnkjfhasdjkfjkladsjflksadjfklsdmlkfjmsld.kfjkml\"}"); err != nil {
		log.Error(ss.Context(), "WriteStringToFile fail", zap.Error(err))
	}
	if err := readFileAndWriteToStream("lgx.txt", ss); err != nil {
		log.Error(ss.Context(), "readFileAndWriteToStream fail", zap.Error(err))
	}
	return nil
}

func readFileAndWriteToStream(filePath string, ss pb.PandaAccount_DownloadFileServer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// 将读取到的内容写入 gRPC 流
		if err := ss.Send(&pb.FileChunk{Data: buffer[:n]}); err != nil {
			return err
		}
	}

	return nil
}
