package roles

import (
	"log"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

func InitCasbin(db *gorm.DB) (*casbin.Enforcer, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		log.Fatalf("Failed to create Casbin adapter: %v", err)
	}

	enforcer, err := casbin.NewEnforcer("../../../storages/model.conf", adapter)
	if err != nil {
		log.Fatalf("Failed to create Casbin enforcer: %v", err)
	}

	if err := enforcer.LoadPolicy(); err != nil {
		log.Fatalf("Failed to load Casbin policy: %v", err)
	}

	return enforcer, nil
}
