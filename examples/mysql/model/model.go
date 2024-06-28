package model

import (
	"context"

	"github.com/asjard/asjard/pkg/database/mysql"
)

func Init() error {
	db, err := mysql.DB(context.Background())
	if err != nil {
		return err
	}
	if err := db.AutoMigrate(&ExampleTable{}); err != nil {
		return err
	}
	return nil
}
