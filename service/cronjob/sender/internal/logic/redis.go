package logic

import "github.com/zeromicro/go-zero/core/stores/redis"

func WithRedis(redisType string, redisPass string) redis.Option {
	return func(p *redis.Redis) {
		p.Type = redisType
		p.Pass = redisPass
	}
}
