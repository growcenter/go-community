package pgsql

import (
	"context"
	"fmt"
	"github.com/lib/pq"
	"go-community/internal/models"
	"go-community/internal/pkg/cursor"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (err error)
	Update(ctx context.Context, user *models.User) (err error)
	UpdateByEmailPhoneNumber(ctx context.Context, email string, phoneNumber string, user *models.User) (err error)
	UpdateByCommunityId(ctx context.Context, communityId string, user *models.User) (err error)
	GetByCommunityId(ctx context.Context, communityId string) (user models.User, err error)
	GetOneByCommunityId(ctx context.Context, communityId string) (user models.User, err error)
	GetByEmail(ctx context.Context, email string) (user models.User, err error)
	GetByPhoneNumber(ctx context.Context, phoneNumber string) (user models.User, err error)
	GetOneByIdentifier(ctx context.Context, identifier string) (user models.User, err error)
	GetOneByEmailPhoneNumber(ctx context.Context, email string, phoneNumber string) (user models.User, err error)
	CheckByEmailPhoneNumber(ctx context.Context, email string, phoneNumber string) (dataExist bool, err error)
	CheckByCommunityId(ctx context.Context, communityId string) (isExist bool, err error)
	GetUserNameByIdentifier(ctx context.Context, identifier string) (output *models.GetNameOnUserDBOutput, err error)
	GetUserNameByCommunityId(ctx context.Context, communityId string) (output *models.GetNameOnUserDBOutput, err error)
	GetAllWithCursor(ctx context.Context, param models.GetAllUserCursorParam) (output []models.GetAllUserDBOutput, prev string, next string, total int, err error)
	BulkUpdateRolesByCommunityIds(ctx context.Context, communityIds []string, roles []string) (err error)
	BulkUpdateUserTypesByCommunityIds(ctx context.Context, communityIds []string, userTypes []string) (err error)
	CheckMultiple(ctx context.Context, communityIds []string) (count int64, err error)
	GetDetailByCommunityId(ctx context.Context, communityId string) (output []models.GetUserProfileDBOutput, err error)
	GetCommunityIdByParams(ctx context.Context, param models.GetCommunityIdsByParameter) (output []models.GetCommunityIdsByParamsDBOutput, err error)
	CountUserByUserTypeCategory(ctx context.Context, userTypeCategory []string) (count int64, err error)
	Delete(ctx context.Context, communityId string) (err error)
	GetRBAC(ctx context.Context, communityId string) (output *models.GetRBACByCommunityIdDBOutput, err error)
}

type userRepository struct {
	db  *gorm.DB
	trx TransactionRepository
}

func NewUserRepository(db *gorm.DB, trx TransactionRepository) UserRepository {
	return &userRepository{db: db, trx: trx}
}

func (ur *userRepository) Create(ctx context.Context, user *models.User) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return ur.trx.Transaction(func(dtx *gorm.DB) error {
		return ur.db.Create(&user).Error
	})
}

func (ur *userRepository) Update(ctx context.Context, user *models.User) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return ur.trx.Transaction(func(dtx *gorm.DB) error {
		return ur.db.Save(&user).Error
	})
}

func (ur *userRepository) UpdateByEmailPhoneNumber(ctx context.Context, email string, phoneNumber string, user *models.User) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	condition, args := ConditionExistOrNot(email, phoneNumber)
	return ur.trx.Transaction(func(dtx *gorm.DB) error {
		return ur.db.Model(&models.User{}).Where(condition, args...).Updates(user).Error
	})
}

func (ur *userRepository) UpdateByCommunityId(ctx context.Context, communityId string, user *models.User) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return ur.db.Model(&models.User{}).Where("community_id = ?", communityId).Updates(user).Error
}

func (ur *userRepository) GetByCommunityId(ctx context.Context, communityId string) (user models.User, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var u models.User
	err = ur.db.Where("community_id = ?", communityId).Find(&u).Error

	return u, err
}

func (ur *userRepository) GetOneByCommunityId(ctx context.Context, communityId string) (user models.User, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var u models.User
	err = ur.db.Where("community_id = ?", communityId).First(&u).Error

	return u, err
}

func (ur *userRepository) GetByEmail(ctx context.Context, email string) (user models.User, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var u models.User
	err = ur.db.Where("email = ?", email).Find(&u).Error

	return u, err
}

func (ur *userRepository) GetByPhoneNumber(ctx context.Context, phoneNumber string) (user models.User, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var u models.User
	err = ur.db.Where("phone_number = ?", phoneNumber).Find(&u).Error

	return u, err
}

func (ur *userRepository) GetOneByIdentifier(ctx context.Context, identifier string) (user models.User, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var u models.User
	err = ur.db.Raw(queryGetOneUserByIdentifier, identifier, identifier).Scan(&u).Error

	return u, err
}

func (ur *userRepository) GetOneByEmailPhoneNumber(ctx context.Context, email string, phoneNumber string) (user models.User, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var u models.User
	err = ur.db.Raw(queryGetOneUserByEmailPhoneNumber, email, email, phoneNumber, phoneNumber).Scan(&u).Error

	return u, err
}

func (ur *userRepository) CheckByEmailPhoneNumber(ctx context.Context, email string, phoneNumber string) (dataExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = ur.db.Raw(queryCheckUserByEmailPhoneNumber, email, email, phoneNumber, phoneNumber).Scan(&dataExist).Error
	if err != nil {
		return false, err
	}

	return dataExist, nil
}

func (ur *userRepository) CheckByCommunityId(ctx context.Context, communityId string) (isExist bool, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = ur.db.Raw(queryCheckUserByCommunityId, communityId).Scan(&isExist).Error
	if err != nil {
		return false, err
	}

	return isExist, nil
}

func (ur *userRepository) GetUserNameByIdentifier(ctx context.Context, identifier string) (output *models.GetNameOnUserDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = ur.db.Raw(queryGetUserNameByIdentifier, identifier).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (ur *userRepository) GetUserNameByCommunityId(ctx context.Context, communityId string) (output *models.GetNameOnUserDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = ur.db.Raw(queryGetUserNameByCommunityId, communityId).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (ur *userRepository) GetAllWithCursor(ctx context.Context, param models.GetAllUserCursorParam) (output []models.GetAllUserDBOutput, prev, next string, total int, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	// Set default limit if none provided
	if param.Limit <= 0 {
		param.Limit = 10 // Default limit
	}

	queryList, paramList, err := BuildQueryGetAllUser(param)
	if err != nil {
		return nil, "", "", 0, fmt.Errorf("failed to build list query: %w", err)
	}

	var records []models.GetAllUserDBOutput
	if err := ur.db.Raw(queryList, paramList...).Scan(&records).Error; err != nil {
		return nil, "", "", 0, fmt.Errorf("query execution failed: %w", err)
	}

	queryCount, paramCount, err := BuildCountGetAllUser(param)
	if err != nil {
		return nil, "", "", 0, fmt.Errorf("failed to build count query: %w", err)
	}

	if err := ur.db.Raw(queryCount, paramCount...).Scan(&total).Error; err != nil {
		return nil, "", "", 0, fmt.Errorf("count query execution failed: %w", err)
	}

	if len(records) > 0 {
		hasMore := len(records) > param.Limit
		isForward := param.Direction != "prev" && param.Cursor != ""
		isBackward := !isForward && param.Direction == "prev" && param.Cursor != ""

		if hasMore {
			records = records[:param.Limit] // Remove the extra record
		}

		if isBackward {
			for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
				records[i], records[j] = records[j], records[i]
			}
		}

		if isBackward || hasMore {
			lastRecord := records[len(records)-1]
			nextCursor := models.GetAllUserCursor{
				CreatedAt: *lastRecord.CreatedAt,
				ID:        lastRecord.ID,
			}
			next = cursor.EncryptCursorFromStruct(nextCursor)
		}

		if isForward || (hasMore && isBackward) {
			firstRecord := records[0]
			prevCursor := models.GetAllUserCursor{
				CreatedAt: *firstRecord.CreatedAt,
				ID:        firstRecord.ID,
			}
			prev = cursor.EncryptCursorFromStruct(prevCursor)
		}
	}

	return records, prev, next, total, nil
}

func (ur *userRepository) BulkUpdateRolesByCommunityIds(ctx context.Context, communityIds []string, roles []string) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	user := models.User{}
	return ur.db.Model(user).Where("community_id IN ?", communityIds).Update("roles", pq.Array(roles)).Error
}

func (ur *userRepository) BulkUpdateUserTypesByCommunityIds(ctx context.Context, communityIds []string, userTypes []string) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	user := models.User{}
	return ur.db.Model(user).Where("community_id IN ?", communityIds).Update("user_types", pq.Array(userTypes)).Error
}

func (ur *userRepository) CheckMultiple(ctx context.Context, communityIds []string) (count int64, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = ur.db.Raw(queryMultipleCheckUser, pq.Array(communityIds)).Scan(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (ur *userRepository) GetDetailByCommunityId(ctx context.Context, communityId string) (output []models.GetUserProfileDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = ur.db.Raw(queryGetProfileByCommunityId, communityId).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (ur *userRepository) GetCommunityIdByParams(ctx context.Context, param models.GetCommunityIdsByParameter) (output []models.GetCommunityIdsByParamsDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	finalQuery, input, err := BuildQueryGetCommunityIdByParams(param)
	if err != nil {
		return nil, err
	}

	err = ur.db.Raw(finalQuery, input...).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (ur *userRepository) CountUserByUserTypeCategory(ctx context.Context, userTypeCategory []string) (count int64, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = ur.db.Raw(queryCountUserByUserTypeCategory, pq.Array(userTypeCategory)).Scan(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (ur *userRepository) Delete(ctx context.Context, communityId string) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return ur.db.Where("community_id = ?", communityId).Delete(&models.User{}).Error
}

func (ur *userRepository) GetRBAC(ctx context.Context, communityId string) (output *models.GetRBACByCommunityIdDBOutput, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	err = ur.db.Raw(queryGetRBACByCommunityId, communityId).Scan(&output).Error
	if err != nil {
		return nil, err
	}

	return output, nil
}
