package db

import (
	"fmt"
	"gorm.io/driver/mysql"
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
	Config, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 构造 DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s",
		Config.Database.User,
		Config.Database.Password,
		Config.Database.Host,
		Config.Database.Port,
		Config.Database.Name,
		Config.Database.Charset,
	)

	// 初始化数据库连接
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}

	log.Println("数据库连接成功")
	SetupConnectionPool()
	//AutoMigrate()
}

// ATInitDB InitDB 初始化数据库连接
func ATInitDB() {
	// 加载配置文件
	Config, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 构造 DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s",
		Config.Database.User,
		Config.Database.Password,
		Config.Database.Host,
		Config.Database.Port,
		Config.Database.Name,
		Config.Database.Charset,
	)

	// 初始化数据库连接
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}

	log.Println("数据库连接成功")
	SetupConnectionPool()
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
