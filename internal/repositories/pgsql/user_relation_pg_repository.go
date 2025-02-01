package pgsql

import (
	"context"
	"go-community/internal/models"
	"gorm.io/gorm"
)

type UserRelationRepository interface {
	Create(ctx context.Context, relation *models.UserRelation) (err error)
	GetOneByRelatedCommunityIds(ctx context.Context, communityId string, relatedCommunityId string) (relation *models.UserRelation, err error)
	Update(ctx context.Context, relation *models.UserRelation) (err error)
	Delete(ctx context.Context, communityId string, relatedCommunityId string) (err error)
}

type userRelationRepository struct {
	db  *gorm.DB
	trx TransactionRepository
}

func NewUserRelationRepository(db *gorm.DB, trx TransactionRepository) UserRelationRepository {
	return &userRelationRepository{db: db, trx: trx}
}

func (urr *userRelationRepository) Create(ctx context.Context, relation *models.UserRelation) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return urr.db.Create(&relation).Error
}

func (urr *userRelationRepository) GetOneByRelatedCommunityIds(ctx context.Context, communityId string, relatedCommunityId string) (relation *models.UserRelation, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = urr.db.Where("community_id = ? AND related_community_id = ?", communityId, relatedCommunityId).Find(&relation).Error
	if err != nil {
		return nil, err
	}

	return relation, nil
}

func (urr *userRelationRepository) Update(ctx context.Context, relation *models.UserRelation) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return urr.db.Save(&relation).Error
}

func (urr *userRelationRepository) Delete(ctx context.Context, communityId string, relatedCommunityId string) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return urr.db.Where("community_id = ? AND related_community_id = ?", communityId, relatedCommunityId).Delete(&models.UserRelation{}).Error
}
