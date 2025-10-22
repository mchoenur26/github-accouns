package services

import ( //library
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

const CacheTTL = time.Hour * 1 //masa berlaku cache
// inisialisasi interaksi antar fungsi
type RedisCacher interface {
	SetClient(slug string, jsonData string) error
	GetClient(slug string) (string, error)
	DeleteClient(slug string) error
}

// inisialisasi redis
type RedisService struct {
	Client *redis.Client
}

func (r *RedisService) SetClient(slug string, jsonData string) error {
	ctx := context.Background()
	return r.Client.Set(ctx, slug, jsonData, CacheTTL).Err()
}

// mengambil data dari database
func (r *RedisService) GetClient(slug string) (string, error) {
	ctx := context.Background() //inisialisasi redi
	val, err := r.Client.Get(ctx, slug).Result()
	//jika terjadi error pada penarikan database atau server
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return val, nil
}

// menghapus data
func (r *RedisService) DeleteClient(slug string) error {
	ctx := context.Background()
	return r.Client.Del(ctx, slug).Err()
}
