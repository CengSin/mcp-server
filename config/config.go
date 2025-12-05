package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
)

type QdrantConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type OpenAIConfig struct {
	BaseURL string `yaml:"baseURL"`
	ApiKey  string `yaml:"apiKey"`
}

type TemporalConfig struct {
	HostPort string `yaml:"hostPort"`
}

type MysqlConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	UserName string `yaml:"userName"`
	Password string `yaml:"password"`
	DB       string `yaml:"DB"`
}

type MinIO struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"accessKeyID"`
	SecretAccessKey string `yaml:"secretAccessKey"`
	BucketName      string `yaml:"bucketName"`
}

type Config struct {
	Qdrant   *QdrantConfig   `yaml:"qdrant"`
	OpenAI   *OpenAIConfig   `yaml:"openAI"`
	Temporal *TemporalConfig `yaml:"temporal"`
	Mysql    *MysqlConfig    `yaml:"mysql"`
	MinIO    *MinIO          `yaml:"minIO"`
}

var (
	Cfg *Config
)

func Parse() {
	Cfg = new(Config)
	if err := cleanenv.ReadConfig("config/config.yaml", Cfg); err != nil {
		log.Fatalln("read config failed, err ", err.Error())
	}
}
