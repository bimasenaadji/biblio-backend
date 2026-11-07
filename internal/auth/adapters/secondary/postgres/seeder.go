package postgres

import (
	"biblio/internal/auth/core/entities"
	"errors"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedRolesAndPermissions populates standard authorization catalog and initial superadmin.
// This function is strictly idempotent: it will safely update or skip existing records.
func SeedRolesAndPermissions(db *gorm.DB) error {
	log.Println("Seeding initial RBAC roles, permissions, and superadmin...")

	// Step 1: Define Standard Granular Permissions
	defaultPermissions := []entities.Permission{
		{Code: "book:read", Description: "View catalog books"},
		{Code: "book:write", Description: "Create, update, and delete catalog books"},
		{Code: "copy:read", Description: "View and lookup physical book copies by barcode"},
		{Code: "copy:write", Description: "Create and update physical book copies"},
		{Code: "loan:create", Description: "Borrow books (circulation desk)"},
		{Code: "loan:return", Description: "Process book returns and audit physical conditions"},
		{Code: "loan:read", Description: "View loan records and monitoring dashboard"},
	}

	// Memory Map for O(1) instant lookup when associating permissions to roles
	permMap := make(map[string]entities.Permission, len(defaultPermissions))

	for _, perm := range defaultPermissions {
		var existing entities.Permission
		err := db.Where("code = ?", perm.Code).
			Attrs(entities.Permission{Description: perm.Description}).
			FirstOrCreate(&existing).Error
		if err != nil {
			return fmt.Errorf("failed to seed permission '%s': %w", perm.Code, err)
		}
		permMap[perm.Code] = existing
	}

	// Step 2: Define Roles and their associated Permission sets
	roleDefs := []struct {
		Name        string
		Description string
		PermCodes   []string
	}{
		{
			Name:        "admin",
			Description: "System Administrator with unrestricted access to all modules",
			PermCodes: []string{
				"book:read", "book:write",
				"copy:read", "copy:write",
				"loan:create", "loan:return", "loan:read",
			},
		},
		{
			Name:        "librarian",
			Description: "Circulation Staff managing inventory and daily borrowing/returns",
			PermCodes: []string{
				"book:read",
				"copy:read", "copy:write",
				"loan:create", "loan:return", "loan:read",
			},
		},
		{
			Name:        "member",
			Description: "Registered Library Member / Student with read access",
			PermCodes: []string{
				"book:read",
				"copy:read",
				"loan:read",
			},
		},
	}

	var adminRole entities.Role

	for _, def := range roleDefs {
		var role entities.Role
		err := db.Where("name = ?", def.Name).
			Attrs(entities.Role{Description: def.Description}).
			FirstOrCreate(&role).Error
		if err != nil {
			return fmt.Errorf("failed to seed role '%s': %w", def.Name, err)
		}

		// Gather permission entities belonging to this role
		var assignedPerms []entities.Permission
		for _, code := range def.PermCodes {
			if p, ok := permMap[code]; ok {
				assignedPerms = append(assignedPerms, p)
			}
		}

		// Sync Many-to-Many relationship in 'role_permissions' pivot table idempotently
		if err := db.Model(&role).Association("Permissions").Replace(assignedPerms); err != nil {
			return fmt.Errorf("failed to associate permissions to role '%s': %w", def.Name, err)
		}

		if def.Name == "admin" {
			adminRole = role
		}
	}

	// Step 3: Seed Default Super Admin Account if not already present
	const defaultAdminEmail = "admin@biblio.local"
	var existingAdmin entities.User

	err := db.Where("email = ?", defaultAdminEmail).First(&existingAdmin).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println("Superadmin user not found, provisioning default admin account...")

		// Generate secure one-way Bcrypt salted hash
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Admin123!"), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash default admin password: %w", err)
		}

		superAdmin := entities.User{
			Email:        defaultAdminEmail,
			PasswordHash: string(hashedPassword),
			FullName:     "System Administrator",
			RoleID:       adminRole.ID,
			IsActive:     true,
		}

		if err := db.Create(&superAdmin).Error; err != nil {
			return fmt.Errorf("failed to create default superadmin: %w", err)
		}
		log.Println("Default superadmin provisioned successfully (email: admin@biblio.local, pass: Admin123!)")
	}

	log.Println("RBAC seeding completed successfully.")
	return nil
}
