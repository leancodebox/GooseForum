package cmd

import (
	"fmt"
	"strconv"

	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/app/service/userservice"
	"github.com/spf13/cobra"
)

func init() {
	appendCommand(&cobra.Command{
		Use:   "set-user-password <userId> <password>",
		Short: "Set a user password",
		Args:  cobra.ExactArgs(2),
		RunE:  runUserSetPassword,
	})
	appendCommand(&cobra.Command{
		Use:   "set-user-email <userId> <email>",
		Short: "Set a user email",
		Args:  cobra.ExactArgs(2),
		RunE:  runUserSetEmail,
	})
	appendCommand(&cobra.Command{
		Use:   "set-user-admin <userId>",
		Short: "Grant the administrator role to a user",
		Args:  cobra.ExactArgs(1),
		RunE:  runUserSetAdmin,
	})
}

func runUserSetPassword(_ *cobra.Command, args []string) error {
	user, err := getUserArg(args[0])
	if err != nil {
		return err
	}
	user.SetPassword(args[1])
	if err := userservice.SaveUser(&user); err != nil {
		return fmt.Errorf("save user password: %w", err)
	}
	fmt.Printf("Password updated for user %d (%s).\n", user.Id, user.Username)
	return nil
}

func runUserSetEmail(_ *cobra.Command, args []string) error {
	user, err := getUserArg(args[0])
	if err != nil {
		return err
	}
	user.Email = args[1]
	if err := userservice.SaveUser(&user); err != nil {
		return fmt.Errorf("save user email: %w", err)
	}
	fmt.Printf("Email updated for user %d (%s): %s\n", user.Id, user.Username, user.Email)
	return nil
}

func runUserSetAdmin(_ *cobra.Command, args []string) error {
	user, err := getUserArg(args[0])
	if err != nil {
		return err
	}

	roleEntity := role.Get(1)
	if roleEntity.Id == 0 {
		roleEntity.RoleName = "管理员"
		roleEntity.Effective = 1
		role.SaveOrCreateById(&roleEntity)
	}

	rs := rolePermissionRs.GetRsByRoleIdAndPermission(roleEntity.Id, permission.Admin.Id())
	if rs.Id == 0 {
		rs.RoleId = roleEntity.Id
		rs.PermissionId = permission.Admin.Id()
		rs.Effective = 1
		rolePermissionRs.SaveOrCreateById(&rs)
	}
	permission.InvalidateRole(roleEntity.Id)

	user.RoleId = roleEntity.Id
	if err := userservice.SaveUser(&user); err != nil {
		return fmt.Errorf("save user role: %w", err)
	}
	fmt.Printf("User %d (%s) is now an administrator.\n", user.Id, user.Username)
	return nil
}

func getUserArg(value string) (users.EntityComplete, error) {
	userID, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return users.EntityComplete{}, fmt.Errorf("invalid user id %q", value)
	}
	user, err := users.Get(userID)
	if err != nil {
		return users.EntityComplete{}, fmt.Errorf("get user %d: %w", userID, err)
	}
	if user.Id == 0 {
		return users.EntityComplete{}, fmt.Errorf("user %d not found", userID)
	}
	return user, nil
}
