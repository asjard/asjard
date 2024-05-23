package main

import (
	"fmt"
	"time"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	apb "github.com/asjard/asjard/examples/protos_repo"

	// _ "github.com/asjard/asjard/pkg/logger"
	_ "github.com/asjard/asjard/pkg/security/base64"
	// _ "github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
)

// Hello .
type Hello struct{}

// Say .
func (h Hello) Say(ctx *rest.Context, in *apb.Empty) (*apb.SayResp, error) {
	fmt.Println("user_id=", in.UserId)
	return &apb.SayResp{Msg: "hello"}, nil
}

// ServiceDesc .
func (h Hello) ServiceDesc() rest.ServiceDesc {
	return apb.HelloRestServiceDesc
}

// // Groups .
// func (Hello) Groups() []*rest.Group {
// 	return []*rest.Group{}
// }

type instance struct {
	ID   string
	Name string
}

type redisConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	DB   int    `yaml:"db"`
}

func main() {
	asjard := asjard.New()
	asjard.AddHandler("rest", &Hello{})
	// asjard.AddHandler("grpc", &Hello{})
	config.Set("test.env.key", "testxxx")
	go func() {
		for {
			// logger.Debugf("http.rest=%s", config.GetString("servers.http.addresses.rest", ""))
			// logger.Debugf("----testEnv=%s", config.GetString("testEnv", "not found"))
			// logger.Debugf("----a=%s", config.GetString("a", "not found"))
			// jsonOut := make(map[string]int)
			// config.GetAndJsonUnmarshal("jsonContent", &jsonOut)
			// logger.Debugf("jsonOut: %+v", jsonOut)
			// yamlOut := make(map[string]int)
			// config.GetAndYamlUnmarshal("yamlContent", &yamlOut)
			// logger.Debugf("yamlOut: %+v", yamlOut)
			// // redisConfig := config.GetWithPrefix("redis")
			// // for key, value := range redisConfig {
			// // 	logger.Debugf("key=%s, value=%v", key, value)
			// // }
			// redisMap := make(map[string]any)
			// config.GetWithUnmarshal("redis", &redisMap)
			// logger.Debugf("redismap: %+v", redisMap)
			// redisStr, err := json.Marshal(&redisMap)
			// logger.Debugf("redisStr: %s, err: %v", string(redisStr), err)
			time.Sleep(3 * time.Second)
			// restAddress := make(map[string]string)
			// config.GetWithUnmarshal("servers.rest.addresses", &restAddress)
			// logger.Debugf("restAddress: %+v", restAddress)
			// logger.Debugf("servers.http.concurrency=%d", config.GetInt("servers.http.concurrency", 10))
		}
	}()
	if err := asjard.Start(); err != nil {
		logger.Error(err.Error())
	}
	// log.Println("----")
	// c := make(chan os.Signal)
	// signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT)
	// select {
	// case s := <-c:
	// 	log.Println("got os signal " + s.String())
	// }
	/*
		{
			"a": {
				"a": {
					"b": [1, 2, 3]
				}
			}
		}
	*/
	logger.Info("exited")
	// services := make(map[string][]*instance)
	// services["x"] = []*instance{{Name: "0"}}
}
