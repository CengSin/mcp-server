package client

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/qdrant/go-client/qdrant"
	"github.com/sashabaranov/go-openai"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"mcp/server/config"
	"os"
)

var (
	Qdrant *qdrant.Client
	AI     *openai.Client
	Mysql  *gorm.DB
	MinIO  *minio.Client
)

func InitMysql(cfg *config.MysqlConfig) {
	// 生成gorm链接配置
	if cfg == nil {
		panic("mysql config is nil")
	}

	// 构建 DSN 连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.UserName,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DB,
	)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             0,           // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,       // Don't include params in the SQL log
			Colorful:                  false,       // Disable color
		},
	)

	// 初始化 MySQL 连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		log.Fatalf("mysql 连接失败: %v", err)
	}

	Mysql = db
	log.Printf("MySQL 初始化成功: host=%s port=%d user=%s db=%s", cfg.Host, cfg.Port, cfg.UserName, cfg.DB)
}

func InitQdrant(cfg *config.QdrantConfig) {
	if cfg == nil {
		panic("qdrant config is nil")
	}

	client, err := qdrant.NewClient(&qdrant.Config{
		Host: cfg.Host,
		Port: cfg.Port,
	})
	if err != nil {
		log.Fatalln("qdrant client init failed, err ", err)
	}

	Qdrant = client
}

func InitLLMs(cfg *config.OpenAIConfig) {
	if cfg == nil {
		panic("openai config is nil")
	}
	if cfg.ApiKey == "" {
		cfg.ApiKey = os.Getenv("OPENROUTER_API_KEY")
	}

	conf := openai.DefaultConfig(cfg.ApiKey)
	if len(cfg.BaseURL) > 0 {
		conf.BaseURL = cfg.BaseURL
	}
	AI = openai.NewClientWithConfig(conf)
}

func InitMinIO(cfg *config.MinIO) {
	var err error
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln("MinIO 连接失败:", err)
	}

	// 自动创建 Bucket (如果不存在)
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		log.Fatalln("检查 Bucket 失败:", err)
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalln("创建 Bucket 失败:", err)
		}
		log.Printf("Bucket '%s' 创建成功\n", cfg.BucketName)
	}

	MinIO = minioClient
}

func Close() {
	Qdrant.Close()
}
