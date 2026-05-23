package auth

// Permission constants — endpoint-level granularity
const (
	PermissionRandom  = "api:random"  // Random image
	PermissionUpload  = "api:upload"  // Upload image
	PermissionImages  = "api:images"  // List images
	PermissionDelete  = "api:delete"  // Delete image
	PermissionTags    = "api:tags"    // Get tags
	PermissionConfig  = "api:config"  // Get config
	PermissionDebug   = "api:debug"   // Debug tags
	PermissionCleanup = "api:cleanup" // Trigger cleanup
)

// AllPermissions is the complete set of available permissions
var AllPermissions = []string{
	PermissionRandom,
	PermissionUpload,
	PermissionImages,
	PermissionDelete,
	PermissionTags,
	PermissionConfig,
	PermissionDebug,
	PermissionCleanup,
}

// Role definitions — each role maps to a set of permissions
var RolePermissions = map[string][]string{
	"reader": {
		PermissionRandom,
		PermissionImages,
		PermissionTags,
		PermissionConfig,
	},
	"writer": {
		PermissionRandom,
		PermissionImages,
		PermissionTags,
		PermissionConfig,
		PermissionUpload,
	},
	"admin": {
		PermissionRandom,
		PermissionImages,
		PermissionTags,
		PermissionConfig,
		PermissionUpload,
		PermissionDelete,
		PermissionCleanup,
		PermissionDebug,
	},
}

// ValidRoles returns the list of valid role names
func ValidRoles() []string {
	roles := make([]string, 0, len(RolePermissions))
	for r := range RolePermissions {
		roles = append(roles, r)
	}
	return roles
}

// IsValidRole checks if a role name is valid
func IsValidRole(role string) bool {
	_, ok := RolePermissions[role]
	return ok
}

// IsValidPermission checks if a permission string is valid
func IsValidPermission(perm string) bool {
	for _, p := range AllPermissions {
		if p == perm {
			return true
		}
	}
	return false
}

// EffectivePermissions returns the effective permission set for a user
// by computing the union of role permissions and custom permissions
func EffectivePermissions(role string, customPermissions []string) []string {
	permSet := make(map[string]bool)

	// Add role permissions
	if rolePerms, ok := RolePermissions[role]; ok {
		for _, p := range rolePerms {
			permSet[p] = true
		}
	}

	// Add custom permissions (union — custom only adds, never subtracts)
	for _, p := range customPermissions {
		if IsValidPermission(p) {
			permSet[p] = true
		}
	}

	result := make([]string, 0, len(permSet))
	for p := range permSet {
		result = append(result, p)
	}
	return result
}

// HasPermission checks if a given permission set includes the required permission
func HasPermission(perms []string, required string) bool {
	for _, p := range perms {
		if p == required {
			return true
		}
	}
	return false
}
