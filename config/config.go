package config

import (
	"github.com/spf13/viper"
)

// Config 应用配置结构体
type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	JWT          JWTConfig
	Site         SiteConfig
	ClientID     string
	ClientSecret string
	Upload       UploadConfig
	GeminiAPIKey string
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int
	Mode string
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret      string
	ExpireHours int `mapstructure:"expire_hours"`
}

// SiteConfig 网站配置
type SiteConfig struct {
	Name        string
	Url         string
	Keywords    string
	Description string
	Version     string
}

// UploadConfig 图床配置
type UploadConfig struct {
	MaxSize       int64    // 最大上传大小(字节)
	AllowedExts   []string // 允许的文件扩展名列表
	ImgurApiUrl   string
	ImgurClientId string // imgur 客户端 ID
}

var config Config

// LoadConfig 加载配置文件
func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = viper.Unmarshal(&config)
	return err
}

// GetConfig 获取配置
func GetConfig() *Config {
	return &config
}
