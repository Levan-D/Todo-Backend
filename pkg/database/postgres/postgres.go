package postgres

import (
	"database/sql"
	"fmt"
	"github.com/Levan-D/Todo-Backend/pkg/config"
	gorm_logrus "github.com/onrik/gorm-logrus"
	migrate "github.com/rubenv/sql-migrate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func NewClient() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Get().Database.Host,
		config.Get().Database.Port,
		config.Get().Database.Username,
		config.Get().Database.Password,
		config.Get().Database.DBName,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gorm_logrus.New()})
	if err != nil {
		return &gorm.DB{}, err
	}

	DB = db

	return db, nil
}

func GetDB() *gorm.DB {
	return DB
}

func MigrationUp() string {
	db, source, err := migrationConnect()
	if err != nil {
		return "cannot be connected to database"
	}
	defer db.Close()

	n, err := migrate.Exec(db, "postgres", source, migrate.Up)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("Applied %d migrations!\n", n)
}

func MigrationDown() string {
	db, source, err := migrationConnect()
	if err != nil {
		return "cannot be connected to database"
	}
	defer db.Close()

	n, err := migrate.ExecMax(db, "postgres", source, migrate.Down, 1)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("Rollback %d migrations!\n", n)
}

func migrationConnect() (*sql.DB, *migrate.PackrMigrationSource, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Get().Database.Host,
		config.Get().Database.Port,
		config.Get().Database.Username,
		config.Get().Database.Password,
		config.Get().Database.DBName,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return &sql.DB{}, &migrate.PackrMigrationSource{}, err
	}

	migrate.SetTable("migration")

	migrations := &migrate.PackrMigrationSource{
		Box: config.GetBox(),
		Dir: "./",
	}

	return db, migrations, nil
}
