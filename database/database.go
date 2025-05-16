package database

import (
	"LinkHUB/models"
	"fmt"
	"log"
	"os"
	"time"

	"LinkHUB/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() error {
	// 获取数据库配置
	dbConfig := config.GetConfig().Database

	// 构建DSN连接字符串
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.SSLMode)
	log.Println(dsn)
	// 配置GORM日志
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second, // 慢SQL阈值
			LogLevel:                  logger.Info, // 日志级别
			IgnoreRecordNotFoundError: true,        // 忽略记录未找到错误
			Colorful:                  true,        // 彩色打印
		},
	)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		return err
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(10)
	// 设置最大打开连接数
	sqlDB.SetMaxOpenConns(100)
	// 设置连接最大生存时间
	sqlDB.SetConnMaxLifetime(time.Hour)
	debugDB := db.Debug() // 开启 Debug 模式的日志
	//err = debugDB.AutoMigrate(&models.Tag{}) // 使用 debugDB 进行迁移

	// 自动迁移数据库表结构
	if err := debugDB.AutoMigrate(
		&models.User{},
		&models.Link{},
		&models.Tag{},
		&models.Vote{},
		&models.Comment{},
		&models.ArticleComment{},
		&models.Article{},
		&models.Category{},
		&models.Notification{},
		&models.Ads{},
		&models.Image{},
	); err != nil {
		return fmt.Errorf("自动迁移数据库表结构失败: %v", err)
	}

	// 赋值给全局变量
	DB = db

	return nil
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return DB
}
