package redis

import (
	"fmt"
	"github.com/go-redis/redis"
	"web_app/settings"
)

var(
	rdb *redis.Client
)

func Init(cfg *settings.RedisConfg) (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%v",cfg.Host,cfg.Port),
		Password: cfg.Password,
		DB: cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	_, err = rdb.Ping().Result()
	return err
}

func Close()  {
	rdb.Close()
}
