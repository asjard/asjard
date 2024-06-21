package snowflake

import (
	"github.com/asjard/asjard/core/bootstrap"
	"github.com/bwmarrin/snowflake"
)

var Node *snowflake.Node

// SnowFlake 雪花ID
type SnowFlake struct{}

func init() {
	bootstrap.AddBootstrap(&SnowFlake{})
}

func (SnowFlake) Bootstrap() error {
	// TODO 生成一个ID
	node, err := snowflake.NewNode(0)
	if err != nil {
		return err
	}
	Node = node
	return nil
}

func (SnowFlake) Shutdown() {}
