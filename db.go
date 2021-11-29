package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func dbOpen() (*gorm.DB, error) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "opensvc"
	}
	pw := os.Getenv("DB_PASSWORD")
	if pw == "" {
		pwf := os.Getenv("DB_PASSWORD_FILE")
		if pwf == "" {
			return nil, fmt.Errorf("One of DB_PASSWORD or DB_PASSWORD_FILE env var is required")
		}
		b, err := ioutil.ReadFile(pwf)
		if err != nil {
			return nil, errors.Wrap(err, "read pwf")
		}
		pw = string(b)
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/opensvc?charset=utf8&parseTime=True&loc=Local", user, pw, host, port)
	return gorm.Open(mysql.New(mysql.Config{
		DSN:               dsn,
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

func initDB() error {
	var err error
	if db, err = dbOpen(); err != nil {
		return errors.Wrap(err, "connect database")
	}
	if err = dbMigrate(db); err != nil {
		return errors.Wrap(err, "migrate database schema")
	}
	return nil
}
