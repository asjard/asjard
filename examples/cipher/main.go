package main

import (
	"context"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/security"
	pb "github.com/asjard/asjard/examples/protobuf/hello"
	cs "github.com/asjard/asjard/pkg/security"
	"github.com/asjard/asjard/pkg/server/rest"
)

const (
	AESCipherName = "aesCBCPkcs5paddingCipherExample"
)

func init() {
	// 添加一个自定义的aes加解密组件
	security.AddCipher(AESCipherName, cs.NewAESCipher)
}

// ExampleCipher 加解密示例
type ExampleCipher struct {
	pb.UnimplementedHelloServer
}

func (ExampleCipher) CipherExample(ctx context.Context, in *pb.CipherExampleReq) (*pb.CipherExampleResp, error) {
	return &pb.CipherExampleResp{
		AesEncryptValueInPlainFile:         config.GetString("testAESEncrptValue", "", config.WithCipher(cs.AESCipherName)),
		Base64EncryptValueInPlainFile:      config.GetString("testBase64EncryptValue", "", config.WithCipher(security.Base64CipherName)),
		PlainValueInAesEncryptFile:         config.GetString("testPlainValueInAESEncryptFile", ""),
		AesEncryptValueInAesEncryptFile:    config.GetString("testAESEncryptValueInAESEncryptFile", "", config.WithCipher(AESCipherName)),
		Base64EncryptValueInAesEncryptFile: config.GetString("testBase64EncryptValueInAESEncryptFile", "", config.WithCipher(security.Base64CipherName)),
	}, nil
}

func (ExampleCipher) RestServiceDesc() *rest.ServiceDesc {
	return &pb.HelloRestServiceDesc
}

func main() {
	server := asjard.New()
	server.AddHandler(&ExampleCipher{}, rest.Protocol)
	if err := server.Start(); err != nil {
		panic(err)
	}
}
