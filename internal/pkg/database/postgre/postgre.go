package postgre

import (
	"fmt"
	"go-community/internal/common"
	"go-community/internal/config"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectWithGORM(config *config.Configuration) (*gorm.DB, error) {
	connection := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s timezone=Asia/Jakarta", config.PostgreSQL.Host, config.PostgreSQL.User, config.PostgreSQL.Password, config.PostgreSQL.Name, config.PostgreSQL.Port, config.PostgreSQL.SSLMode)
	fmt.Sprintln(connection)
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  connection,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().In(common.GetLocation())
		},
	})

	if err != nil {
		return nil, err
	}

	// Set the timezone
	tx := db.Exec("SET timezone TO 'Asia/Jakarta'")
	if tx.Error != nil {
		panic("failed to set timezone: " + tx.Error.Error())
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
