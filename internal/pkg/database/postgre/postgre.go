package postgre

import (
	"fmt"
	"go-community/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectWithGORM(config *config.Configuration) (*gorm.DB, error) {
	connection := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Jakarta", config.PostgreSQL.Host, config.PostgreSQL.User, config.PostgreSQL.Password, config.PostgreSQL.Name, config.PostgreSQL.Port, config.PostgreSQL.SSLMode)
	fmt.Sprintln(connection)
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  connection,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	return db, nil
}

func CloseGORM(db *gorm.DB) error {
	pg, err := db.DB()
	if err != nil {
		return err
	}
	return pg.Close()
}
