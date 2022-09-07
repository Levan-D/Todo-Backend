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
			"config": "/etc/todo",
			"log":    "/var/log/todo",
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
