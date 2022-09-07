package config

import (
	"fmt"
	"github.com/Levan-D/Todo-Backend/pkg/logger"
	"github.com/Levan-D/Todo-Backend/pkg/utils"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"time"
)

var (
	IsDevelopment = false
)

var (
	config      *Config
	directories map[string]string
)

type Config struct {
	Server struct {
		Schema string `yaml:"schema"`
		Domain string `yaml:"domain"`
		Host   string `yaml:"host"`
		Port   uint16 `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     uint16 `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
	} `yaml:"database"`
	RabbitMQ struct {
		Host     string `yaml:"host"`
		Port     uint16 `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"rabbitmq"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     uint16 `yaml:"port"`
		Password string `yaml:"password"`
	} `yaml:"redis"`
	JWT struct {
		AccessSecret  string        `yaml:"accessSecret"`
		RefreshSecret string        `yaml:"refreshSecret"`
		AccessTTL     time.Duration `yaml:"accessTTL"`
		RefreshTTL    time.Duration `yaml:"refreshTTL"`
	} `yaml:"jwt"`
	SMTP struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		From     string `yaml:"from"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"smtp"`
	AWS struct {
		S3 struct {
			Region          string `yaml:"region"`
			AccessKeyID     string `yaml:"accessKeyId"`
			SecretAccessKey string `yaml:"secretAccessKey"`
			Bucket          string `yaml:"bucket"`
		} `yaml:"aws"`
	} `yaml:"aws"`
	BranchIO struct {
		Key         string `yaml:"key"`
		Secret      string `yaml:"secret"`
		AppID       string `yaml:"appId"`
		AccessToken string `yaml:"accessToken"`
	} `yaml:"branchio"`
	TBCCheckout struct {
		ClientID     string `yaml:"clientId"`
		ClientSecret string `yaml:"clientSecret"`
		APIKey       string `yaml:"apiKey"`
	} `yaml:"tbcCheckout"`
	Gateway struct {
		Firebase struct {
			ProjectID   string `yaml:"projectId"`
			DatabaseURL string `yaml:"databaseUrl"`
			APIKey      string `yaml:"apiKey"`
		} `yaml:"firebase"`
		Facebook struct {
			ClientID     string `yaml:"clientId"`
			ClientSecret string `yaml:"clientSecret"`
		} `yaml:"facebook"`
		Google struct {
			ClientID        string `yaml:"clientId"`
			ClientSecret    string `yaml:"clientSecret"`
			ClientIOSID     string `yaml:"clientIosId"`
			ClientAndroidID string `yaml:"clientAndroidId"`
		} `yaml:"google"`
		Apple struct {
			TeamID   string `yaml:"teamId"`
			KeyID    string `yaml:"keyId"`
			ClientID string `yaml:"clientId"`
		} `yaml:"apple"`
	} `yaml:"gateway"`
}

func init() {
	_ = godotenv.Load(".env")

	if os.Getenv("MODE") == "development" {
		IsDevelopment = true

		pwd, err := os.Getwd()
		if err != nil {
			log.WithFields(log.Fields{
				"package":  "pkg/config",
				"function": "init",
				"error":    err,
			}).Fatal("cannot get os.Getwd result")
		}

		directories = map[string]string{
			"config": pwd + "/configs",
			"log":    pwd + "/.log",
			"tmp":    pwd + "/.tmp",
		}

	} else {
		IsDevelopment = false

		directories = map[string]string{
			"config": "/etc/loyalty",
			"log":    "/var/log/loyalty",
			"tmp":    "/tmp",
		}
	}

	createDirectories()
}

func Get() *Config {
	return config
}

func Initialize(module string) {
	logger.SetFileSaveData(module, GetDirectoryPath("log"))

	// Setup
	viper.SetConfigName(fmt.Sprintf("config.%s", module))
	viper.SetConfigType("yaml")
	viper.AddConfigPath(GetDirectoryPath("config"))

	err := viper.ReadInConfig()

	if err != nil {
		log.WithFields(log.Fields{
			"package":  "pkg/config",
			"function": "Initialize",
			"error":    err,
		}).Fatal("error while reading config file: ", err)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.WithFields(log.Fields{
			"package":  "pkg/config",
			"function": "Initialize",
			"error":    err,
		}).Fatal("unable to decode into config struct, %v", err)
	}

	setConfigFromEnv(config)
}

func setConfigFromEnv(config *Config) {
	if os.Getenv("SERVER_HOST") != "" {
		config.Server.Host = os.Getenv("SERVER_HOST")
	}
	if os.Getenv("SERVER_PORT") != "" {
		n, err := strconv.ParseUint(os.Getenv("SERVER_PORT"), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		config.Server.Port = uint16(n)
	}
	if os.Getenv("DATABASE_HOST") != "" {
		config.Database.Host = os.Getenv("DATABASE_HOST")
	}
	if os.Getenv("DATABASE_PORT") != "" {
		n, err := strconv.ParseUint(os.Getenv("DATABASE_PORT"), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		config.Database.Port = uint16(n)
	}
	if os.Getenv("DATABASE_USERNAME") != "" {
		config.Database.Username = os.Getenv("DATABASE_USERNAME")
	}
	if os.Getenv("DATABASE_PASSWORD") != "" {
		config.Database.Password = os.Getenv("DATABASE_PASSWORD")
	}
	if os.Getenv("RABBITMQ_HOST") != "" {
		config.RabbitMQ.Host = os.Getenv("RABBITMQ_HOST")
	}
	if os.Getenv("RABBITMQ_PORT") != "" {
		n, err := strconv.ParseUint(os.Getenv("RABBITMQ_PORT"), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		config.RabbitMQ.Port = uint16(n)
	}
	if os.Getenv("RABBITMQ_USERNAME") != "" {
		config.RabbitMQ.Username = os.Getenv("RABBITMQ_USERNAME")
	}
	if os.Getenv("RABBITMQ_PASSWORD") != "" {
		config.RabbitMQ.Password = os.Getenv("RABBITMQ_PASSWORD")
	}
	if os.Getenv("DATABASE_DBNAME") != "" {
		config.Database.DBName = os.Getenv("DATABASE_DBNAME")
	}
	if os.Getenv("REDIS_HOST") != "" {
		config.Redis.Host = os.Getenv("REDIS_HOST")
	}
	if os.Getenv("REDIS_PORT") != "" {
		n, err := strconv.ParseUint(os.Getenv("REDIS_PORT"), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		config.Redis.Port = uint16(n)
	}
	if os.Getenv("REDIS_PASSWORD") != "" {
		config.Redis.Password = os.Getenv("REDIS_PASSWORD")
	}
	if os.Getenv("JWT_ACCESS_SECRET") != "" {
		config.JWT.AccessSecret = os.Getenv("JWT_ACCESS_SECRET")
	}
	if os.Getenv("JWT_REFRESH_SECRET") != "" {
		config.JWT.RefreshSecret = os.Getenv("JWT_REFRESH_SECRET")
	}
	if os.Getenv("JWT_ACCESS_TTL") != "" {
		t, err := time.ParseDuration(os.Getenv("JWT_ACCESS_TTL"))
		if err != nil {
			log.Fatal(err)
		}
		config.JWT.AccessTTL = t
	}
	if os.Getenv("JWT_REFRESH_TTL") != "" {
		t, err := time.ParseDuration(os.Getenv("JWT_REFRESH_TTL"))
		if err != nil {
			log.Fatal(err)
		}
		config.JWT.RefreshTTL = t
	}
	if os.Getenv("AWS_S3_REGION") != "" {
		config.AWS.S3.Region = os.Getenv("AWS_S3_REGION")
	}
	if os.Getenv("AWS_S3_ACCESS_KEY_ID") != "" {
		config.AWS.S3.AccessKeyID = os.Getenv("AWS_S3_ACCESS_KEY_ID")
	}
	if os.Getenv("AWS_S3_SECRET_ACCESS_KEY") != "" {
		config.AWS.S3.SecretAccessKey = os.Getenv("AWS_S3_SECRET_ACCESS_KEY")
	}
	if os.Getenv("AWS_S3_BUCKET") != "" {
		config.AWS.S3.Bucket = os.Getenv("AWS_S3_BUCKET")
	}
	if os.Getenv("GATEWAY_FIREBASE_PROJECT_ID") != "" {
		config.Gateway.Firebase.ProjectID = os.Getenv("GATEWAY_FIREBASE_PROJECT_ID")
	}
	if os.Getenv("GATEWAY_FIREBASE_DATABASE_URL") != "" {
		config.Gateway.Firebase.DatabaseURL = os.Getenv("GATEWAY_FIREBASE_DATABASE_URL")
	}
	if os.Getenv("GATEWAY_FIREBASE_API_KEY") != "" {
		config.Gateway.Firebase.APIKey = os.Getenv("GATEWAY_FIREBASE_API_KEY")
	}
	if os.Getenv("GATEWAY_FACEBOOK_CLIENT_ID") != "" {
		config.Gateway.Facebook.ClientID = os.Getenv("GATEWAY_FACEBOOK_CLIENT_ID")
	}
	if os.Getenv("GATEWAY_FACEBOOK_CLIENT_SECRET") != "" {
		config.Gateway.Facebook.ClientSecret = os.Getenv("GATEWAY_FACEBOOK_CLIENT_SECRET")
	}
	if os.Getenv("GATEWAY_GOOGLE_CLIENT_ID") != "" {
		config.Gateway.Google.ClientID = os.Getenv("GATEWAY_GOOGLE_CLIENT_ID")
	}
	if os.Getenv("GATEWAY_GOOGLE_CLIENT_SECRET") != "" {
		config.Gateway.Google.ClientSecret = os.Getenv("GATEWAY_GOOGLE_CLIENT_SECRET")
	}
	if os.Getenv("GATEWAY_GOOGLE_CLIENT_IOS_ID") != "" {
		config.Gateway.Google.ClientIOSID = os.Getenv("GATEWAY_GOOGLE_CLIENT_IOS_ID")
	}
	if os.Getenv("GATEWAY_GOOGLE_CLIENT_ANDROID_ID") != "" {
		config.Gateway.Google.ClientAndroidID = os.Getenv("GATEWAY_GOOGLE_CLIENT_ANDROID_ID")
	}
	if os.Getenv("GATEWAY_APPLE_TEAM_ID") != "" {
		config.Gateway.Apple.TeamID = os.Getenv("GATEWAY_APPLE_TEAM_ID")
	}
	if os.Getenv("GATEWAY_APPLE_KEY_ID") != "" {
		config.Gateway.Apple.KeyID = os.Getenv("GATEWAY_APPLE_KEY_ID")
	}
	if os.Getenv("GATEWAY_APPLE_CLIENT_ID") != "" {
		config.Gateway.Apple.ClientID = os.Getenv("GATEWAY_APPLE_CLIENT_ID")
	}
}

func GetDirectoryPath(dirKey string) string {
	for key, value := range directories {
		if key == dirKey {
			return value
		}
	}

	log.WithFields(log.Fields{
		"package":  "pkg/config",
		"function": "GetDirectoryPath",
	}).Fatal("cannot found directory key: ", dirKey)

	return ""
}

func createDirectories() {
	for _, value := range directories {
		err := utils.CreateDirectoryIfNotExists(value)
		if err != nil {
			log.WithFields(log.Fields{
				"package":  "pkg/config",
				"function": "createDirectories",
				"error":    err,
			}).Fatal("cannot be created folder: ", value)
		}
	}
}
