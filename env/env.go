package env

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Env struct {
	Env   		string 	  `env:"ENV" env-required:"true"`
	Http  		Http   	  `env-required:"true"`
	PgSql 		PgSql  	  `env-required:"true"`
	MinIO 		MinIO  	  `env-required:"true"`
	Redis 		Redis     `env-required:"true"`
	AdminToken 	string 	  `env:"ADMIN_TOKEN" env-required:"true"`
	JWTSecret 	string    `env:"JWT_SECRET" env-required:"true"`
}

type Http struct {
	Host string `env:"HOST" env-required:"true"`
	Port string `env:"PORT" env-required:"true"`
}

type PgSql struct {
	Host     string `env:"POSTGRES_HOST" env-required:"true"`
	User     string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DbName   string `env:"POSTGRES_DB" env-required:"true"`
	Port     int    `env:"POSTGRES_PORT" env-default:"5432"`
	SSLMode  string `env:"POSTGRES_SSLMODE" env-default:"disable"`
	URI      string `env:"POSTGRES_URI"`
}

type MinIO struct {
	Endpoint               string `env:"MINIO_ENDPOINT" env-default:"localhost:9000"`
	AccessKey              string `env:"MINIO_ACCESS_KEY" env-default:"minioadmin"`
	SecretKey              string `env:"MINIO_SECRET_KEY" env-default:"minioadmin"`
	UseSSL                 bool   `env:"MINIO_USE_SSL" env-default:"false"`
	BucketName             string `env:"MINIO_BUCKET_NAME" env-default:"documents"`
	Region                 string
	CreateBucketIfNotExist bool `env:"MINIO_CREATE_BUCKET_IF_NOT_EXIST" env-default:"true"`
}

type Redis struct {
	Addr string `env:"REDIS_ADDR" env-required:"true"`
	Pass string `env:"REDIS_PASS" env-required:"true"`
}

func (c *Http) GetPort() string {
	if envPort := os.Getenv("PORT"); envPort != "" {
		return envPort
	}
	return c.Port
}

func MustLoad() *Env {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: No .env file found")
	}

	var config Env

	if err := cleanenv.ReadEnv(&config); err != nil {
		panic("failed to read env: " + err.Error())
	}

	return &config
}