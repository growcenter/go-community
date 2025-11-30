package pgsql

import (
	"context"
	"go-community/internal/models"

	"gorm.io/gorm"
)

type AccessRelationshipRepository interface {
	Create(ctx context.Context, accessRelationship *models.AccessRelationship) (err error)
	Delete(ctx context.Context, accessRelationship *models.AccessRelationship) (err error)
	GetOneById(ctx context.Context, id string) (accessRelationship *models.AccessRelationship, err error)
	// GetRecursiveRelationships fetches all relationships relevant to a permission check in a single query.
	// It uses a recursive SQL query to traverse the relationship graph, starting from the initial object
	// and including relationships of parent objects (e.g., the community that owns an event).
	// This is the core optimization to avoid multiple database round-trips.
	GetRecursiveRelationships(ctx context.Context, objectType string, objectID string) ([]*models.AccessRelationship, error)
}

type accessRelationshipRepository struct {
	db *gorm.DB
}

func NewAccessRelationshipRepository(db *gorm.DB) AccessRelationshipRepository {
	return &accessRelationshipRepository{db: db}
}

func (arr *accessRelationshipRepository) Create(ctx context.Context, accessRelationship *models.AccessRelationship) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return arr.db.WithContext(ctx).Create(accessRelationship).Error
}

func (arr *accessRelationshipRepository) Delete(ctx context.Context, accessRelationship *models.AccessRelationship) (err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return arr.db.WithContext(ctx).Delete(accessRelationship).Error
}

func (arr *accessRelationshipRepository) GetOneById(ctx context.Context, id string) (accessRelationship *models.AccessRelationship, err error) {
	defer func() {
		LogRepository(ctx, err)
	}()

	return accessRelationship, arr.db.WithContext(ctx).First(&accessRelationship, "id = ?", id).Error
}

// GetRecursiveRelationships uses a recursive Common Table Expression (CTE) to fetch the entire
// subgraph of relationships relevant to a given object. This is highly efficient as it
// performs the graph traversal inside the database in a single query.
func (arr *accessRelationshipRepository) GetRecursiveRelationships(ctx context.Context, objectType string, objectID string) ([]*models.AccessRelationship, error) {
	defer func() {
		LogRepository(ctx, nil) // Note: err is handled manually below
	}()

	var relationships []*models.AccessRelationship

	// This raw SQL query is the heart of the performance optimization.
	// It builds a dependency graph for a given object and returns all tuples in it.
	query := `
        WITH RECURSIVE relevant_rels AS (
            -- Base case: Start with the direct relationships on the target object
            SELECT *
            FROM access_relationships
            WHERE object_type = ? AND object_id = ? AND deleted_at IS NULL

            UNION ALL

            -- Recursive step: Find relationships connected to the subjects of the previous level.
            -- This traverses the graph upwards (e.g., from event -> community -> user).
            SELECT ar.*
            FROM access_relationships ar
            JOIN relevant_rels rr ON ar.object_type = rr.subject_type AND ar.object_id = rr.subject_id
            WHERE ar.deleted_at IS NULL
        )
        SELECT * FROM relevant_rels;
    `

	err := arr.db.WithContext(ctx).Raw(query, objectType, objectID).Scan(&relationships).Error
	if err != nil {
		LogRepository(ctx, err)
		return nil, err
	}

	return relationships, nil
}
