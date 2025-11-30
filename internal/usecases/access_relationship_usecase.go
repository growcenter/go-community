package usecases

import (
	"context"
	"go-community/internal/models"
	"go-community/internal/pkg/authz"
	"go-community/internal/repositories/pgsql"
)

// AccessRelationshipUsecase defines the interface for managing access relationships and checking permissions.
// It orchestrates the interaction between the database repository and the pure authorization logic.
type AccessRelationshipUsecase interface {
	// CheckPermission verifies if a subject has a given permission on a resource.
	// This is the primary method for performing authorization checks.
	CheckPermission(ctx context.Context, subjectType, subjectID, permission, objectType, objectID string) (bool, error)

	// WriteRelationship creates a new relationship between a subject and a resource.
	// This is called when application state changes (e.g., a user joins a group).
	WriteRelationship(ctx context.Context, subjectType, subjectID, relation, objectType, objectID string) error

	// DeleteRelationship removes an existing relationship.
	DeleteRelationship(ctx context.Context, subjectType, subjectID, relation, objectType, objectID string) error
}

type accessRelationshipUsecase struct {
	repo       pgsql.AccessRelationshipRepository
	authzModel *authz.Schema
}

// NewAccessRelationshipUsecase creates a new AccessRelationshipUsecase.
func NewAccessRelationshipUsecase(repo pgsql.AccessRelationshipRepository) AccessRelationshipUsecase {
	return &accessRelationshipUsecase{
		repo:       repo,
		authzModel: authz.NewAuthzModel(), // Initialize the authorization model.
	}
}

// CheckPermission implements the core authorization logic.
// It now uses a single, efficient call to the repository to get all relevant relationships.
func (uc *accessRelationshipUsecase) CheckPermission(ctx context.Context, subjectType, subjectID, permission, objectType, objectID string) (bool, error) {
	// Fetch the entire relevant relationship graph in one database call.
	// This is the critical performance optimization.
	allRels, err := uc.repo.GetRecursiveRelationships(ctx, objectType, objectID)
	if err != nil {
		return false, err
	}

	// Convert the database models to the simplified Relationship struct required by the authz package.
	// This decoupling prevents the pure authz logic from depending on database models.
	logicRels := make([]authz.Relationship, len(allRels))
	for i, r := range allRels {
		logicRels[i] = authz.Relationship{
			ObjectType:  r.ObjectType,
			ObjectID:    r.ObjectId,
			Relation:    r.Relation,
			SubjectType: r.SubjectType,
			SubjectID:   r.SubjectId,
		}
	}

	// Call the pure Check function from the authz package to get the final answer.
	// The checking logic itself remains the same, but it now operates on the full dataset.
	return authz.Check(uc.authzModel, logicRels, subjectType, subjectID, permission, objectType, objectID), nil
}

// WriteRelationship creates and persists a new access relationship.
func (uc *accessRelationshipUsecase) WriteRelationship(ctx context.Context, subjectType, subjectID, relation, objectType, objectID string) error {
	newRel := &models.AccessRelationship{
		SubjectType: subjectType,
		SubjectId:   subjectID,
		Relation:    relation,
		ObjectType:  objectType,
		ObjectId:    objectID,
	}
	return uc.repo.Create(ctx, newRel)
}

// DeleteRelationship removes an access relationship.
// Note: This is a simple implementation that assumes a unique relationship.
// A more robust version would find the specific relationship to delete.
func (uc *accessRelationshipUsecase) DeleteRelationship(ctx context.Context, subjectType, subjectID, relation, objectType, objectID string) error {
	rel := &models.AccessRelationship{
		SubjectType: subjectType,
		SubjectId:   subjectID,
		Relation:    relation,
		ObjectType:  objectType,
		ObjectId:    objectID,
	}
	return uc.repo.Delete(ctx, rel)
}
