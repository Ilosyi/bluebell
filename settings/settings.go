package settings

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// GlobalConfig 全局配置
var GlobalConfig *Config

// Config 配置结构体
type Config struct {
	App   AppConfig   `mapstructure:"app"`
	Log   LogConfig   `mapstructure:"log"`
	MySQL MySQLConfig `mapstructure:"mysql"`
	Redis RedisConfig `mapstructure:"redis"`
}

type AppConfig struct {
	Name      string `mapstructure:"name"`
	Mode      string `mapstructure:"mode"`
	Port      int    `mapstructure:"port"`
	StartTime string `mapstructure:"start_time"`
	MachineID int64  `mapstructure:"machine_id"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	DB   int    `mapstructure:"db"`
}

// Init 初始化配置
func Init() (err error) {
	configPath := "./settings/config.yaml"
	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		configPath = "./settings/config.example.yaml"
	}

	viper.SetConfigFile(configPath)
	err = viper.ReadInConfig()
	if err != nil {
		fmt.Printf("viper.ReadInConfig() failed with %s\n", err)
		return
	}

	err = viper.Unmarshal(&GlobalConfig)
	if err != nil {
		fmt.Printf("viper.Unmarshal() failed with %s\n", err)
		return
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		viper.Unmarshal(&GlobalConfig)
	})
	return
}
