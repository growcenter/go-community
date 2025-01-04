package pgsql

import (
	"context"
	"fmt"
	"go-community/internal/models"
	"go-community/internal/pkg/cursor"
	"gorm.io/gorm"
	"strconv"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (err error)
	Update(ctx context.Context, user *models.User) (err error)
	UpdateByEmailPhoneNumber(ctx context.Context, email string, phoneNumber string, user *models.User) (err error)
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

func (ur *userRepository) GetAllWithCursor(ctx context.Context, param models.GetAllUserCursorParam) (output []models.GetAllUserDBOutput, prev string, next string, total int, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	var decryptedId int64
	var totalEntries int

	// Decrypt cursor if provided (based on updated_at)
	if param.Cursor != "" {
		decryptedId, err = cursor.DecryptCursorFromInteger(param.Cursor)
		if err != nil {
			return nil, "", "", 0, err
		}
	}

	// Set default limit if none provided
	limit := param.Limit
	if limit <= 0 {
		limit = 10 // Default limit
	}

	// Build the query
	query, params, err := BuildQueryGetAllUser(baseQueryGetAllUser, param.SearchBy, param.Search, param.CampusCode, param.CoolId, param.Department, decryptedId, param.Direction, limit)
	if err != nil {
		return nil, "", "", 0, err
	}

	fmt.Println("query: ", query)
	fmt.Println("param: ", query)

	// Apply limit (fetch `limit + 1` to check for extra entries)
	//params[len(params)-1] = limit + 1

	// Execute query
	var records []models.GetAllUserDBOutput
	err = ur.db.Raw(query, params...).Scan(&records).Error
	if err != nil {
		return nil, "", "", 0, err
	}

	// Get total count for pagination info
	countQuery, countParams, _ := BuildQueryGetAllUser(queryCountAllUser, param.SearchBy, param.Search, param.CampusCode, param.CoolId, param.Department, decryptedId, param.Direction, limit)
	err = ur.db.Raw(countQuery, countParams...).Scan(&totalEntries).Error
	if err != nil {
		return nil, "", "", 0, err
	}

	// Handle pagination logic
	hasMore := len(records) == limit && len(records) < totalEntries
	if hasMore {
		records = records[:limit] // Trim to the limit
	}

	// Generate cursors
	if len(records) > 0 {
		// Generate a `prev` cursor for non-first pages
		if param.Cursor != "" {
			prev, err = cursor.EncryptCursor(strconv.Itoa(records[0].ID))
			if err != nil {
				return nil, "", "", 0, err
			}
		}

		// Generate a `next` cursor if there are more entries
		if hasMore {
			next, err = cursor.EncryptCursor(strconv.Itoa(records[len(records)-1].ID))
			if err != nil {
				return nil, "", "", 0, err
			}
		}
	}

	// Return results
	return records, prev, next, totalEntries, nil
}
