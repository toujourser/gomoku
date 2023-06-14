package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"github.com/toujourser/gomoku/pkg/logger"
	"net"
	"time"
)

var (
	RedisClient *redis.Client
)

func init() {
	RedisClient = redis.NewClient(&redis.Options{
		//连接信息
		Network:  "tcp",                         //网络类型，tcp or unix，默认tcp
		Addr:     viper.GetString("redis.addr"), //主机名+冒号+端口，默认localhost:6379
		Password: "",                            //密码
		DB:       0,                             // redis数据库index

		//连接池容量及闲置连接数量
		PoolSize:     15, // 连接池最大socket连接数，默认为4倍CPU数， 4 * runtime.NumCPU
		MinIdleConns: 10, //在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量；。

		//超时
		DialTimeout:  5 * time.Second, //连接建立超时时间，默认5秒。
		ReadTimeout:  3 * time.Second, //读超时，默认3秒， -1表示取消读超时
		WriteTimeout: 3 * time.Second, //写超时，默认等于读超时
		PoolTimeout:  4 * time.Second, //当所有连接都处在繁忙状态时，客户端等待可用连接的最大等待时长，默认为读超时+1秒。

		//命令执行失败时的重试策略
		MaxRetries:      0,                      //命令执行失败时，最多重试多少次，默认为0即不重试
		MinRetryBackoff: 8 * time.Millisecond,   //每次计算重试间隔时间的下限，默认8毫秒，-1表示取消间隔
		MaxRetryBackoff: 512 * time.Millisecond, //每次计算重试间隔时间的上限，默认512毫秒，-1表示取消间隔

		//可自定义连接函数
		Dialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
			netDialer := &net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Minute,
			}
			return netDialer.Dial("tcp", addr)
		},

		//钩子函数
		OnConnect: func(ctx context.Context, conn *redis.Conn) error { //仅当客户端执行命令时需要从连接池获取连接时，如果连接池需要新建连接时则会调用此钩子函数
			logger.Infof("conn=%s", conn)
			return nil
		},
	})
	if RedisClient.Ping(context.Background()).Err() != nil {
		panic(RedisClient.Ping(context.Background()).Err())
	}

}
