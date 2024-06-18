package model

import "github.com/asjard/asjard/pkg/database/mysql"

func Init() error {
	db, err := mysql.DB()
	if err != nil {
		return err
	}
	if err := db.AutoMigrate(&ExampleTable{}); err != nil {
		return err
	}
	return nil
}
