package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func dbOpen() (*gorm.DB, error) {
	// DB_SLOW_QUERY_THRESHOLD
	slowQueryThresholdStr := os.Getenv("DB_SLOW_QUERY_THRESHOLD")
	slowQueryThreshold, err := time.ParseDuration(slowQueryThresholdStr)
	if err != nil {
		slowQueryThreshold = time.Second
	}

	// DB_LOG_LEVEL
	logLvlStr := os.Getenv("DB_LOG_LEVEL")
	logLvl := logger.Silent
	switch logLvlStr {
	case "info":
		logLvl = logger.Info
	case "warn", "warning":
		logLvl = logger.Warn
	case "err", "error":
		logLvl = logger.Error
	}

	// DB_HOST
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	// DB_PORT
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}

	// DB_USER
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "opensvc"
	}

	// DB_PASSWORD, DB_PASSWORD_FILE
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

	newLogger := logger.New(
		log.New(os.Stdout, "", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             slowQueryThreshold, // Slow SQL threshold
			LogLevel:                  logLvl,             // Log level
			IgnoreRecordNotFoundError: true,               // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,               // Disable color
		},
	)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/opensvc?charset=utf8&parseTime=True&loc=Local", user, pw, host, port)
	return gorm.Open(mysql.New(mysql.Config{
		DSN:               dsn,
		DefaultStringSize: 256,
		//DisableDatetimePrecision:  true,
		//DontSupportRenameIndex:    true,
		//DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		Logger: newLogger,
	})
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
