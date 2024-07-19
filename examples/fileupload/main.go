package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/examples/protobuf/filepb"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

type FileAPI struct {
	*filepb.FileAPI
	savePath string
}

func NewFileAPI() *FileAPI {
	return &FileAPI{
		FileAPI:  &filepb.FileAPI{},
		savePath: os.TempDir(),
	}
}

// Upload 文件上传
func (api *FileAPI) Upload(ctx context.Context, in *filepb.UploadReq) (*filepb.UploadResp, error) {
	rtx, ok := ctx.(*rest.Context)
	if !ok {
		return nil, status.UnsupportProtocol
	}
	fh, err := rtx.FormFile("file")
	if err != nil {
		logger.Error("read file fail", "err", err)
		return nil, status.Error(codes.NotFound, "file not found")
	}
	fileName := uuid.NewString() + filepath.Ext(fh.Filename)
	if err := fasthttp.SaveMultipartFile(fh, filepath.Join(api.savePath, fileName)); err != nil {
		logger.Error("save file fail", "err", err)
		return nil, status.InternalServerError
	}

	return &filepb.UploadResp{
		DownloadName: fileName,
	}, nil
}

// Download 文件下载
func (api *FileAPI) Download(ctx context.Context, in *filepb.DownloadReq) (*emptypb.Empty, error) {
	rtx, ok := ctx.(*rest.Context)
	if !ok {
		return nil, status.UnsupportProtocol
	}
	filePath := filepath.Join(api.savePath, in.FileName)
	if !utils.IsPathExists(filePath) {
		return nil, status.Errorf(codes.NotFound, "file %s not found", in.FileName)
	}
	rtx.Response.Header.Set(fasthttp.HeaderContentDisposition, "attachment;filename="+in.FileName)
	rtx.SendFile(filePath)
	return nil, nil
}

func main() {
	server := asjard.New()
	server.AddHandler(NewFileAPI(), rest.Protocol)
	if err := server.Start(); err != nil {
		panic(err)
	}
}
