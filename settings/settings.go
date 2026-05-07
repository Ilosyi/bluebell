package settings

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// GlobalConfig 保存整个项目运行时的全局配置。
// 其他包通过 settings.GlobalConfig 读取配置，例如 MySQL 地址、Redis 地址、JWT 密钥。
// 注意：它在 settings.Init 成功后才会有值，因此 main.go 必须最先调用 settings.Init。
var GlobalConfig *Config

// Config 对应 settings/config.yaml 的顶层结构。
// mapstructure tag 告诉 viper：yaml 里的 app/log/mysql/redis/jwt/rate_limit 分别映射到哪个字段。
type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Log       LogConfig       `mapstructure:"log"`
	MySQL     MySQLConfig     `mapstructure:"mysql"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

// AppConfig 保存应用自身配置。
// StartTime 和 MachineID 会传给雪花算法，用于生成业务 ID。
type AppConfig struct {
	Name          string `mapstructure:"name"`
	Mode          string `mapstructure:"mode"`
	Port          int    `mapstructure:"port"`
	StartTime     string `mapstructure:"start_time"`
	MachineID     int64  `mapstructure:"machine_id"`
	EnableSwagger bool   `mapstructure:"enable_swagger"`
}

// LogConfig 保存日志配置。
// MaxSize/MaxAge/MaxBackups 用于控制日志文件切割和保留策略。
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

// MySQLConfig 保存 MySQL 连接配置和连接池配置。
// MaxOpenConns/MaxIdleConns 可以避免高并发时无限创建数据库连接。
type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// RedisConfig 保存 Redis 地址、密码与数据库编号。
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig 保存 JWT 签名密钥和过期时间。
// Secret 用于签名和验签，泄露后攻击者可能伪造 token。
type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	ExpireSeconds int64  `mapstructure:"expire_seconds"`
}

// RateLimitConfig 保存全局令牌桶限流配置。
//
// Enabled 用来控制是否启用限流，开发调试时如果觉得频率太低，可以临时改成 false。
// Rate 表示“每秒往桶里放多少个令牌”，例如 100 表示平均每秒允许 100 个请求通过。
// Capacity 表示“桶最多能存多少个令牌”，容量越大，越能容忍短时间突发流量。
//
// 令牌桶可以这样理解：
// 1. 桶里有令牌，请求拿走 1 个令牌后继续执行。
// 2. 桶里没有令牌，请求被限流。
// 3. 后台会按 Rate 的速度持续补充令牌，但最多补到 Capacity。
type RateLimitConfig struct {
	Enabled  bool    `mapstructure:"enabled"`
	Rate     float64 `mapstructure:"rate"`
	Capacity int64   `mapstructure:"capacity"`
}

// Init 初始化配置。
// 调用流程：
// 1. 如果 BLUEBELL_CONFIG_FILE 环境变量存在，就优先读取它指定的配置文件。
// 2. 否则优先使用 settings/config.yaml。
// 3. 如果本地配置不存在，则回退到 settings/config.example.yaml。
// 4. 使用 viper 读取 yaml，并反序列化到 GlobalConfig。
// 5. 监听配置文件变更，变更后自动重新加载。
func Init() (err error) {
	// Docker/生产环境通常不希望把配置文件固定成 settings/config.yaml。
	// 因此这里先读取环境变量，让容器可以通过 BLUEBELL_CONFIG_FILE 指向专用配置。
	configPath := os.Getenv("BLUEBELL_CONFIG_FILE")
	if configPath == "" {
		// 默认读取本地开发配置。
		configPath = "./settings/config.yaml"
		// 如果本地配置不存在，则读取示例配置，方便首次启动。
		if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
			configPath = "./settings/config.example.yaml"
		}
	}

	// 告诉 viper 配置文件的路径。
	viper.SetConfigFile(configPath)
	// 真正读取配置文件内容。
	err = viper.ReadInConfig()
	if err != nil {
		fmt.Printf("viper.ReadInConfig() failed with %s\n", err)
		return
	}

	// 把 yaml 内容映射到 Config 结构体。
	err = viper.Unmarshal(&GlobalConfig)
	if err != nil {
		fmt.Printf("viper.Unmarshal() failed with %s\n", err)
		return
	}

	// 监听配置文件变化。开发时修改配置后，不需要重启进程即可更新 GlobalConfig。
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		// 重新解析配置。这里没有中断程序，适合开发环境热更新配置。
		viper.Unmarshal(&GlobalConfig)
	})
	return
}
