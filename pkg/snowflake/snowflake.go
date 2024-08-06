package snowflake

import (
	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/runtime"
	"github.com/bwmarrin/snowflake"
)

var Node *snowflake.Node

// SnowFlake 雪花ID
type SnowFlake struct{}

func init() {
	bootstrap.AddBootstrap(&SnowFlake{})
}

func (SnowFlake) Bootstrap() error {
	// TODO 添加实例ID
	node, err := snowflake.NewNode(int64(runtime.GetAPP().Instance.SystemCode))
	if err != nil {
		return err
	}
	Node = node
	return nil
}

func (SnowFlake) Shutdown() {}
