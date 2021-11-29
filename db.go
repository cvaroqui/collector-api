package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func dbOpen() (*gorm.DB, error) {
	return gorm.Open(mysql.New(mysql.Config{
		DSN:               "opensvc:opensvc4@tcp(127.0.0.1:3306)/opensvc?charset=utf8&parseTime=True&loc=Local",
		DefaultStringSize: 256,
		//DisableDatetimePrecision:  true,
		//DontSupportRenameIndex:    true,
		//DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{})
}

func dbMigrate(db *gorm.DB) error {
	if err := db.Table("auth_node").AutoMigrate(&authNode{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&Node{}); err != nil {
		return err
	}
	return nil
}

func init() {
	var err error
	db, err = dbOpen()
	if err != nil {
		panic("failed to connect database")
	}
	if err := dbMigrate(db); err != nil {
		panic("failed to migrate schema")
	}
}
