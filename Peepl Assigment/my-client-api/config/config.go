package config

import ( //library
	"log"
	"os"
)

type Config struct { //menentukan struktur data
	DatabaseURL string
	RedisAddr   string
	RedisPass   string
	AWSRegion   string
	S3Bucket    string
}

// inisialisasi konfigurasi sitem
func LoadConfig() *Config {
	//memberi informasi bahwa sistem sedan memuat
	log.Println("Loading application configuration...")
	//inisialisasi koneksi dengan PostgreSQL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/go_products?sslmode=disable"
		log.Println("WARNING: DATABASE_URL not set, using default development connection.")
	}
	//inisialisasi identitas redis yang akan digunakan
	redisAddr := os.Getenv("REDIS_ADDR") //alamat website
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	} //konfirmasi password
	redisPass := os.Getenv("REDIS_PASSWORD")
	awsRegion := os.Getenv("AWS_REGION") //region
	if awsRegion == "" {
		awsRegion = "ap-southeast-1"
	} //inisialisasi koneksi dengan S3
	s3Bucket := os.Getenv("S3_BUCKET_NAME")
	if s3Bucket == "" {
		log.Fatal("FATAL: S3_BUCKET_NAME is required for the application to run.")
	}
	//menarik nilai dalam database
	return &Config{
		DatabaseURL: dbURL,
		RedisAddr:   redisAddr,
		RedisPass:   redisPass,
		AWSRegion:   awsRegion,
		S3Bucket:    s3Bucket,
	}
}
