package usecases

import "go-community/internal/repositories/pgsql"

type Dependencies struct {
	Repository	*pgsql.PostgreRepositories
}

type Usecases struct {
	Health	healthUsecase
	Campus	campusUsecase
	CoolCategory coolCategoryUsecase
}

func New(d Dependencies) *Usecases{
	health := NewHealthUsecase(d.Repository.Health)
	campus := NewCampusUsecase(d.Repository.Campus)
	coolCategory := NewCoolDivisionUsecase(d.Repository.CoolCategory)

	return &Usecases{
		Health: *health,
		Campus: *campus,
		CoolCategory: *coolCategory,
	}
}