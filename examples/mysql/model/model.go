package model

import (
	"context"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/database/mysql"
)

func Init() error {
	logger.Debug("create table")
	db, err := mysql.DB(context.Background())
	if err != nil {
		return err
	}
	if err := db.AutoMigrate(&ExampleTable{}); err != nil {
		return err
	}
	return nil
}
