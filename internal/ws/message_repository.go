package ws

import (
	"context"
	_ "context"
	_ "fmt"
	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	Redis *redis.Client
	ctx   context.Context
}

func NewRedisRepository() RedisRepository {
	return RedisRepository{Redis: redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}),
		ctx: context.Background(),
	}
}

func (r RedisRepository) SetData(key string, value interface{}) {
	r.Redis.Set(r.ctx, key, value, 0)
}
func (r RedisRepository) GetData(key string) *redis.StringCmd {
	return r.Redis.Get(r.ctx, key)
}

func (r RedisRepository) SetMessage(roomId string, messageId string, message Message) *redis.BoolCmd {
	return r.Redis.HMSet(r.ctx, roomId, messageId, message)
}
