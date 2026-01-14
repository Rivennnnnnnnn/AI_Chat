package utils

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

func InitSnowflake(nodeTitle int64) error {
	// 设置一个起始时间（纪元），一旦确定不可更改
	snowflake.Epoch = time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC).UnixNano() / 1000000
	var err error
	node, err = snowflake.NewNode(nodeTitle) // 0-1023 之间的机器 ID
	return err
}

func GenerateID() int64 {
	return node.Generate().Int64()
}
