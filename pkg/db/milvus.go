package db

import (
	"AI_Chat/pkg/utils"
	"context"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
)

var MilvusClient client.Client

func InitMilvus(cfg utils.MilvusConfig) error {
	if cfg.Address == "" {
		return nil
	}
	cli, err := client.NewClient(context.Background(), client.Config{
		Address:  cfg.Address,
		Username: cfg.Username,
		Password: cfg.Password,
		DBName:   cfg.Database,
	})
	if err != nil {
		return err
	}
	MilvusClient = cli
	return nil
}
