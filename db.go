package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db     *gorm.DB
	tables map[string]*table = map[string]*table{}
)

func dbOpen() (*gorm.DB, error) {
	// API_DB_SLOW_QUERY_THRESHOLD
	slowQueryThresholdStr := viper.GetString("db.log.slow_query_threshold")
	slowQueryThreshold, err := time.ParseDuration(slowQueryThresholdStr)
	if err != nil {
		slowQueryThreshold = time.Second
	}

	// API_DB_LOG_LEVEL
	logLvlStr := viper.GetString("db.log.level")
	logLvl := logger.Silent
	switch logLvlStr {
	case "info":
		logLvl = logger.Info
	case "warn", "warning":
		logLvl = logger.Warn
	case "err", "error":
		logLvl = logger.Error
	}

	// API_DB_HOST, API_DB_PORT, API_DB_USERNAME
	host := viper.GetString("db.host")
	port := viper.GetString("db.port")
	user := viper.GetString("db.username")

	// API_DB_PASSWORD, API_DB_PASSWORD_FILE
	pw := viper.GetString("db.password")
	if pw == "" {
		pwf := viper.GetString("db.password_file")
		if pwf == "" {
			return nil, fmt.Errorf("One of API_DB_PASSWORD or API_DB_PASSWORD_FILE env var is required")
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

func dbMigrate() error {
	for _, t := range tables {
		if err := t.AutoMigrate(); err != nil {
			return err
		}
	}
	return nil
}

func initDB() error {
	var err error
	if db, err = dbOpen(); err != nil {
		return errors.Wrap(err, "connect database")
	}
	if err = dbMigrate(); err != nil {
		return errors.Wrap(err, "migrate database schema")
	}
	return nil
}
