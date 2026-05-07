// Package logger 封装项目的日志初始化和 Gin 中间件。
// 这个包的核心目标：
// 1. 初始化全局 zap logger，方便任何包通过 zap.L() 记录日志。
// 2. 记录每个 HTTP 请求的基本信息。
// 3. 捕获 panic，避免单个请求把整个服务打崩。
package logger

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"bluebell/settings"
)

// Init 根据配置初始化全局 zap logger。
// cfg 来自配置文件中的 log 节点，mode 来自 app.mode。
// 开发模式下日志输出到控制台；非开发模式下日志输出到文件。
func Init(cfg *settings.LogConfig, mode string) (err error) {
	if settings.GlobalConfig == nil {
		return
	}

	// 把配置文件里的字符串级别（debug/info/warn/error）解析成 zap 内部级别。
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return err
	}

	// lumberjack 负责日志文件切割。
	// 例如文件超过 MaxSize MB 后自动切一个新文件，旧文件按 MaxAge/MaxBackups 清理。
	logWriter := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
	}

	// EncoderConfig 决定日志每个字段叫什么，以及时间、级别、调用者等如何格式化。
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// JSONEncoder 适合生产环境：结构化日志便于机器检索。
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	var core zapcore.Core
	if mode == "dev" {
		// 开发模式用更易读的控制台格式，直接输出到 stdout。
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core = zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zap.DebugLevel)
	} else {
		// 生产模式把 JSON 日志写入文件，并使用配置指定的日志级别。
		core = zapcore.NewCore(
			encoder,
			zapcore.AddSync(logWriter),
			zap.NewAtomicLevelAt(level),
		)
	}

	// AddCaller 会记录日志所在的文件和行号。
	// AddCallerSkip(1) 用于跳过一层封装，让日志定位更接近业务调用点。
	globalLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	// 替换 zap 的全局 logger，之后任意包都可以用 zap.L() 获取。
	zap.ReplaceGlobals(globalLogger)

	return nil
}

// GinLogger 是 Gin 请求日志中间件。
// 每个请求都会记录：状态码、方法、路径、查询参数、客户端 IP、User-Agent、耗时等。
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间，用来计算请求耗时。
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// c.Next 会执行后续中间件和真正的 handler。
		// handler 执行完后，才继续往下记录响应状态和耗时。
		c.Next()

		cost := time.Since(start)
		zap.L().Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

// GinRecovery 捕获项目中可能出现的 panic。
// 如果没有 recover，某个 handler panic 可能导致服务进程崩溃。
// stack=true 时会把调用栈写入日志，方便定位 panic 发生位置。
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// defer 会在当前函数返回前执行；recover 只能在 defer 中捕获 panic。
		defer func() {
			if err := recover(); err != nil {
				// broken pipe / connection reset 通常表示客户端已经断开连接。
				// 这类错误不一定是服务端 bug，所以单独识别并减少无意义堆栈。
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				// DumpRequest 把 HTTP 请求转成文本，写日志时用于排查问题。
				// 第二个参数 false 表示不打印请求 body，避免日志过大或泄露敏感数据。
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					zap.L().Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// 客户端连接已经断开，无法再写 HTTP 响应，只能终止请求链。
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				// 根据调用方是否需要 stack，选择是否记录完整调用栈。
				if stack {
					zap.L().Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					zap.L().Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				// 给客户端返回 500，表示服务器内部错误。
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		// 执行后续 handler；如果后续发生 panic，会回到上面的 defer 中处理。
		c.Next()
	}
}
