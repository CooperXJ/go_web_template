package settings

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

//全局变量，用来保存所有程序的配置
var Config = new(AppConfig)

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Mode    string `mapstructure:"mode"`
	Version string `mapstructure:"version"`
	Port    int    `mapstructure:"port"`

	*LogConfig   `mapstructure:"log"`
	*MySQLConfig `mapstructure:"mysql"`
	*RedisConfg  `mapstructure:"redis"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	FileName   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Port         int    `mapstructure:"port"`
	DbName       string `mapstructure:"dbname"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfg struct {
	Host     string `mapstructure:"host"`
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

func Init() (err error) {
	//读取配置文件  不带有文件后缀
	viper.SetConfigName("config")

	//设置文件类型 一般是配合远程配置中心
	//viper.SetConfigType("yaml")

	viper.AddConfigPath("./conf") //还可以在工作目录中查找配置

	err = viper.ReadInConfig() //查找并读取配置文件
	if err != nil {
		panic(fmt.Errorf("fatal err config file %s\n", err))
	}

	//读取配置并反序列化到config中
	err = viper.Unmarshal(Config)
	if err != nil {
		fmt.Println("反序列化失败")
		return err
	}

	fmt.Printf("c: %#v\n", Config)

	//实时监控配置文件的变化
	viper.WatchConfig()
	//当配置变化之后调用一个回调函数
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("config file changed", in.Name)
		//如果配置文件发送了变化，那么会自动更新结构体Config
		if err := viper.Unmarshal(Config); err != nil {
			fmt.Printf("反序列化失败")
		}
	})

	return nil
}
