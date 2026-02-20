package cache

import "fmt"

// Cache key generators for consistent key naming

// RolesCacheKey generates a cache key for roles validation
func RolesCacheKey(roleIDs []string) string {
	return fmt.Sprintf("roles:%v", roleIDs)
}

// UserTypesCacheKey generates a cache key for user types validation
func UserTypesCacheKey(userTypeIDs []string) string {
	return fmt.Sprintf("user_types:%v", userTypeIDs)
}

// CommunityCacheKey generates a cache key for community validation
func CommunityCacheKey(communityID string) string {
	return fmt.Sprintf("community:%s", communityID)
}

// EventCodeCacheKey generates a cache key for event code existence check
func EventCodeCacheKey(code string) string {
	return fmt.Sprintf("event_code:%s", code)
}

// InstanceCodeCacheKey generates a cache key for instance code existence check
func InstanceCodeCacheKey(code string) string {
	return fmt.Sprintf("instance_code:%s", code)
}
