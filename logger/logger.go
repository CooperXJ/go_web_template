package logger

import (
	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"
	"web_app/settings"
)

//Init 初始化日志
func Init(cfg *settings.LogConfig) (err error) {
	logWriter := getLogWriter(cfg)
	encoder := getEncoder()

	//因为这里的日志level是zap内的一个类型，因此需要反序列化
	var l = new(zapcore.Level)
	err = l.UnmarshalText([]byte(viper.GetString("log.level")))
	if err != nil {
		return err
	}

	core := zapcore.NewCore(encoder, logWriter,l)

	//替换zap库中的全局logger
	logger := zap.New(core,zap.AddCaller())  //zap.AddCaller() 记录哪一行调用了日志打印信息
	zap.ReplaceGlobals(logger)
	return
}


/**
日志配置  如何在gin中使用zap日志记录呢？r:=gin.New() r.Use(Zap日志的MiddleWare,GinRecovery(logger,true))  将原来gin中的日志去除掉，然后在middleWare中添加日志
*/

//getLogWriter 获取文件句柄 指定将日志写到哪里
func getLogWriter(cfg *settings.LogConfig) zapcore.WriteSyncer{
	//启用日志分隔
	lumberJackLogger := &lumberjack.Logger{
		Filename:   cfg.FileName,
		MaxSize:    cfg.MaxSize,    //以M为单位  一个日志文件大小
		MaxAge:     cfg.MaxAge,    //以天为单位 保留备份的时间
		MaxBackups: cfg.MaxBackups,     // 最大备份数量  就是当日志文件数量超过MaxSize之后生成的备份文件数量
		LocalTime:  false, //启用本地时间
		Compress:   false, //是否压缩
	}

	//file, _ := os.OpenFile("./test.log",os.O_CREATE|os.O_APPEND|os.O_RDWR,0744)
	return zapcore.AddSync(lumberJackLogger)
}

//getEncoder 获取编码器 如何写入日志 是使用json格式编码还是像console一样输出流就可以了
func getEncoder() zapcore.Encoder{
	//输出日志配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
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
	//json格式
	return zapcore.NewJSONEncoder(encoderConfig)
}

//GinLogger 将zap作为gin的logger
func GinLogger() gin.HandlerFunc {
	return func(context *gin.Context) {
		start:=time.Now()
		path:=context.Request.URL.Path
		query:=context.Request.URL.RawQuery

		context.Next()

		cost:=time.Since(start)
		zap.L().Info(path,
			zap.Int("status",context.Writer.Status()),
			zap.String("method",context.Request.Method),
			zap.String("path",path),
			zap.String("query",query),
			zap.String("ip",context.ClientIP()),
			zap.String("user-agent",context.Request.UserAgent()),
			zap.String("errors",context.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost",cost),
		)
	}
}

// GinRecovery recover掉项目可能出现的panic
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					zap.L().Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

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
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}


