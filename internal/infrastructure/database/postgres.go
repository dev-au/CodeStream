package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/dev-au/CodeStream/internal/config"
	"github.com/dev-au/CodeStream/pkg/logs"
)

func NewPostgresClient(cfg *config.PostgreSQLConfig) (*gorm.DB, func(), error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logs.NewGormLogger(),
	})
	if err != nil {
		return nil, nil, err
	}

	cleanupMySQL := func() {
		sqlDB, err := db.DB()
		if err == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
	}

	return db, cleanupMySQL, nil
}
