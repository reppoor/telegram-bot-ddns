package db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"telegrambot/config"
	"telegrambot/internal/db/models"
	"time"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() {
	// 加载配置文件
	Config, err := config.LoadConfig("") // 加载配置路径
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	var dsn string

	// 根据配置文件选择 MySQL 或 SQLite
	if Config.Database.Type == "mysql" {
		// 构造 MySQL DSN
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s",
			Config.Database.User,
			Config.Database.Password,
			Config.Database.Host,
			Config.Database.Port,
			Config.Database.Name,
			Config.Database.Charset,
		)
	} else if Config.Database.Type == "sqlite" {
		// 构造 SQLite DSN
		dsn = Config.Database.File // SQLite 使用文件路径
	} else {
		log.Fatalf("不支持的数据库类型: %v", Config.Database.Type)
	}

	// 初始化数据库连接
	var db *gorm.DB
	if Config.Database.Type == "mysql" {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else if Config.Database.Type == "sqlite" {
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}

	DB = db
	log.Println("数据库连接成功")

	// 设置连接池等（适用于 MySQL，SQLite 没有这么复杂）
	if Config.Database.Type == "mysql" {
		SetupConnectionPool()
	}
}

// ATInitDB InitDB 初始化数据库连接
func ATInitDB() {
	// 加载配置文件
	Config, err := config.LoadConfig("") // 加载配置路径
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	var dsn string

	// 根据配置文件选择 MySQL 或 SQLite
	if Config.Database.Type == "mysql" {
		// 构造 MySQL DSN
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s",
			Config.Database.User,
			Config.Database.Password,
			Config.Database.Host,
			Config.Database.Port,
			Config.Database.Name,
			Config.Database.Charset,
		)
	} else if Config.Database.Type == "sqlite" {
		// 构造 SQLite DSN
		dsn = Config.Database.File // SQLite 使用文件路径
	} else {
		log.Fatalf("不支持的数据库类型: %v", Config.Database.Type)
	}

	// 初始化数据库连接
	var db *gorm.DB
	if Config.Database.Type == "mysql" {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else if Config.Database.Type == "sqlite" {
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}

	DB = db
	log.Println("数据库连接成功")

	// 设置连接池等（适用于 MySQL，SQLite 没有这么复杂）
	if Config.Database.Type == "mysql" {
		SetupConnectionPool()
	}
	AutoMigrate()
}

// SetupConnectionPool 配置连接池
func SetupConnectionPool() {
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("获取数据库连接失败: %v", err)
	}

	sqlDB.SetMaxOpenConns(1000)
	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
}

// AutoMigrate 自动迁移数据库模型
func AutoMigrate() {
	err := DB.AutoMigrate(
		// 添加你的数据模型
		&models.Domain{},
		&models.TelegramPermission{},
	)
	if err != nil {
		log.Fatalf("自动迁移失败: %v", err)
	}
	log.Println("数据库模型迁移成功")
}

// CloseDB 关闭数据库连接
func CloseDB() {
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("获取数据库连接失败: %v", err)
	}
	err = sqlDB.Close()
	if err != nil {
		log.Printf("关闭数据库连接失败: %v", err)
	} else {
		log.Println("数据库连接已成功关闭")
	}
}
