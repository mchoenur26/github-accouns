package main

import ( //library
	"context"
	"log"

	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"

	localCfg "my-client-api/config"
	"my-client-api/handlers"
	"my-client-api/services"
)

func main() { //koneksi .ENV
	if err := godotenv.Load(); err != nil {
		log.Println("WARNING: No .env file found or unable to load. Relying on system environment variables.")
	}
	//memanggil config local
	cfg := localCfg.LoadConfig()
	//koneksi dengan PostgreSQL
	dbPool, err := pgxpool.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Gagal koneksi ke PostgreSQL:", err)
	}
	defer dbPool.Close()
	//koneksi dengan redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
	})
	redisService := services.RedisService{Client: rdb}

	//koneksi dan konfigurasi AWSS3
	cfgAws, err := awsCfg.LoadDefaultConfig(context.TODO(), awsCfg.WithRegion(cfg.AWSRegion))
	if err != nil {
		log.Fatal("Gagal memuat konfigurasi AWS:", err)
	}
	s3Client := s3.NewFromConfig(cfgAws)

	s3Service := services.S3Service{
		Client:     s3Client,
		BucketName: cfg.S3Bucket,
	}
	//framework fiber
	app := fiber.New()
	//inisialisasi handlerclient
	clientHandler := handlers.ClientHandler{
		DB:    dbPool,
		Redis: &redisService,
		S3:    &s3Service,
	}
	api := app.Group("/clients")
	//fungsi CRUD
	api.Get("/:slug", clientHandler.GetClientBySlug)
	api.Post("/", clientHandler.CreateClient)
	api.Delete("/:slug", clientHandler.DeleteClient)
	//menjalankan local server
	log.Fatal(app.Listen(":3000"))
}
