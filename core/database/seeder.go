package database

import (
	"fmt"
	"template-go/base/helpers"
	"template-go/data/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func initSeeder(DB *gorm.DB) (err error) {

	return DB.Transaction(func(tx *gorm.DB) error {
		// 1. Privileges
		privileges := []string{
			"FilterUser",
			"DeleteUser",
			"RegisterAdminUser",
			"InsertPrivilegeGroup",
			"AssignPrivilegeToGroup",
			"ChangeAdminPrivilegeRole",
			"CheckAdminGroupPrivilegeExists",
			"FilterPrivilegeGroup",
			"FilterPrivilege",
		}

		superUserGroup := "SuperUser"
		superUserUsername := "superuser"

		for _, p := range privileges {
			priv := model.Privilege{Privilege: p}
			if err := tx.Where(model.Privilege{Privilege: p}).FirstOrCreate(&priv).Error; err != nil {
				return fmt.Errorf("seed privilege %q: %w", p, err)
			}
		}

		// 2. Admin privilege group (role)
		group := model.PrivilegeGroup{Code: superUserGroup, Description: "Administrator"}
		if err := tx.Where(model.PrivilegeGroup{Code: superUserGroup}).FirstOrCreate(&group).Error; err != nil {
			return fmt.Errorf("seed privilege group: %w", err)
		}

		// 3. Link privileges to the admin group
		for _, p := range privileges {
			link := model.PrivilegeGroupPrivilege{PrivilegeGroupCode: superUserGroup, Privilege: p}
			if err := tx.Where(model.PrivilegeGroupPrivilege{
				PrivilegeGroupCode: superUserGroup,
				Privilege:          p,
			}).FirstOrCreate(&link).Error; err != nil {
				return fmt.Errorf("seed group privilege %q: %w", p, err)
			}
		}

		// 4. Super user
		password := "toor"
		hashed, err := helpers.HashPassword(password)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}

		now, err := helpers.TimeNow()
		if err != nil {
			return fmt.Errorf("time now: %w", err)
		}
		user := model.User{
			ExternalID: uuid.NewString(),
			Username:   superUserUsername,
			Email:      "superuser@admin.com",
			Password:   hashed,
			FullName:   "Super User",
			Role:       "Admin",
			Timestamp: &model.Timestamp{
				CreatedTs:     now,
				LastUpdatedTs: now,
			},
		}
		// Only created if no user with this username exists; an existing user is left untouched.
		if err := tx.Where(model.User{Username: superUserUsername}).FirstOrCreate(&user).Error; err != nil {
			return fmt.Errorf("seed super user: %w", err)
		}

		// 5. Assign the user to the admin group (user_id is unique, so one group per user)
		userGroup := model.UserPrivilegeGroup{UserID: user.ID, PrivilegeGroupCode: superUserGroup}
		if err := tx.Where(model.UserPrivilegeGroup{UserID: user.ID}).FirstOrCreate(&userGroup).Error; err != nil {
			return fmt.Errorf("assign user to group: %w", err)
		}

		// 6. create new user role.
		userRole := model.Role{
			Code:        "User",
			Description: "Normal User",
		}
		if err := tx.Where(model.Role{Code: "User"}).FirstOrCreate(&userRole).Error; err != nil {
			return fmt.Errorf("create user role: %w", err)
		}

		return nil
	})
}
