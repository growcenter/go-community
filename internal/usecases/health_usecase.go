package usecases

import (
	"context"
	"go-community/internal/repositories/pgsql"
)

type HealthUsecase interface {
	Check(ctx context.Context) (err error)
}

type healthUsecase struct {
	hr	pgsql.HealthRepository
}

func NewHealthUsecase(hr pgsql.HealthRepository) *healthUsecase {
	return &healthUsecase{
		hr: hr,
	}
}

func (hu *healthUsecase) Check(ctx context.Context) (err error) {
	return hu.hr.Check(ctx)
}

// type HealthUsecase interface {
// 	Health(ctx context.Context) map[string]string
// }

// type health usecase

// var _ HealthUsecase = (*health)(nil)

// func (hu *health) Health(ctx context.Context) map[string]string {
// 	status := map[string]string{
// 		"psql": "postgresql is up and running",
// 	}

// 	if err := hu.u.postgreRepository.Ping(ctx); err != nil {
// 		status["psql"] = fmt.Sprintf("failed connect to mysql: %v", err)
// 	}

// 	return status
// }
