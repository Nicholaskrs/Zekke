package service_test

import (
	"context"
	"net/http"
	dto "template-go/base/base"
	"template-go/base/helpers"
	permissionSvc "template-go/modules/permission/svc"
	"template-go/modules/user/permission"
	userSvc "template-go/modules/user/svc"
	testassert "template-go/test"
	"template-go/test/fixture"
	"testing"
)

func TestUserAdminService_RegisterNewAdmin(t *testing.T) {
	t.Parallel()

	type input struct {
		UserAdminId         int
		AdminPrivilegeGroup string
		Username            string
		Email               string
		Password            string
		FullName            string
		PrivilegeGroup      string
	}
	type expected struct {
		success          bool
		errorCode        int
		assignCalls      int
		privilegeCalls   int
		groupExistsCalls int
	}

	type testCase struct {
		name                      string
		input                     input
		expected                  expected
		privilegeExistsMock       *permissionSvc.CheckAdminGroupPrivilegeExistsOut
		hasPrivilegeMock          *permissionSvc.HasPrivilegeOut
		changeAdminPermissionMock *permissionSvc.ChangeAdminPrivilegeRoleOut
	}

	// +----------+
	// | Username |
	// +----------+
	// | taken    |
	// +----------+
	// Default value for mock value.
	defaultPrivilegeExistsMock := func() *permissionSvc.CheckAdminGroupPrivilegeExistsOut {
		return &permissionSvc.CheckAdminGroupPrivilegeExistsOut{
			BaseOut: dto.BaseOut{
				Success:      true,
				ErrorMessage: "",
				ErrorCode:    0,
			},
			IsExists: true,
		}
	}
	defaultHasPrivilegeMock := func() *permissionSvc.HasPrivilegeOut {
		return &permissionSvc.HasPrivilegeOut{
			BaseOut: dto.BaseOut{
				Success:      true,
				ErrorMessage: "",
				ErrorCode:    0,
			},
			IsAllowed: true,
		}
	}
	defaultChangeAdminPermissionMock := func() *permissionSvc.ChangeAdminPrivilegeRoleOut {
		return &permissionSvc.ChangeAdminPrivilegeRoleOut{
			BaseOut: dto.BaseOut{
				Success:      true,
				ErrorMessage: "",
				ErrorCode:    0,
			},
		}
	}

	tests := []testCase{
		{
			name: "registers admin and assigns requested group",
			input: input{
				UserAdminId:         10,
				AdminPrivilegeGroup: "SuperAdmin",
				Username:            "new-admin",
				Email:               "new-admin@gmail.com",
				Password:            "helloWorld",
				FullName:            "John Smith",
				PrivilegeGroup:      "Auditor",
			},
			expected: expected{
				success:          true,
				assignCalls:      1,
				privilegeCalls:   1,
				groupExistsCalls: 1,
				errorCode:        0,
			},
			privilegeExistsMock:       defaultPrivilegeExistsMock(),
			hasPrivilegeMock:          defaultHasPrivilegeMock(),
			changeAdminPermissionMock: defaultChangeAdminPermissionMock(),
		},
		{
			name: "rejects empty admin user ID",
			input: input{
				UserAdminId:         0,
				AdminPrivilegeGroup: "SuperAdmin",
				Username:            "new-admin2",
				Email:               "new-admin2@gmail.com",
				Password:            "helloWorld2",
				FullName:            "John Smith2",
				PrivilegeGroup:      "Auditor",
			},
			expected: expected{
				success:          false,
				errorCode:        http.StatusBadRequest,
				assignCalls:      0,
				privilegeCalls:   0,
				groupExistsCalls: 0,
			},
			privilegeExistsMock:       defaultPrivilegeExistsMock(),
			hasPrivilegeMock:          defaultHasPrivilegeMock(),
			changeAdminPermissionMock: defaultChangeAdminPermissionMock(),
		},
		{
			name: "rejects empty admin privilege group",
			input: input{
				UserAdminId:         10,
				AdminPrivilegeGroup: "",
				Username:            "new-admin2",
				Email:               "new-admin2@gmail.com",
				Password:            "helloWorld2",
				FullName:            "John Smith2",
				PrivilegeGroup:      "Auditor",
			},
			expected: expected{
				success:          false,
				errorCode:        http.StatusBadRequest,
				assignCalls:      0,
				privilegeCalls:   0,
				groupExistsCalls: 0,
			},
			privilegeExistsMock:       defaultPrivilegeExistsMock(),
			hasPrivilegeMock:          defaultHasPrivilegeMock(),
			changeAdminPermissionMock: defaultChangeAdminPermissionMock(),
		},
		{
			name: "rejects empty username",
			input: input{
				UserAdminId:         10,
				AdminPrivilegeGroup: "SuperAdmin",
				Username:            "",
				Email:               "new-admin2@gmail.com",
				Password:            "helloWorld2",
				FullName:            "John Smith2",
				PrivilegeGroup:      "Auditor",
			},
			expected: expected{
				success:          false,
				errorCode:        http.StatusBadRequest,
				assignCalls:      0,
				privilegeCalls:   0,
				groupExistsCalls: 0,
			},
			privilegeExistsMock:       defaultPrivilegeExistsMock(),
			hasPrivilegeMock:          defaultHasPrivilegeMock(),
			changeAdminPermissionMock: defaultChangeAdminPermissionMock(),
		},
		{
			name: "rejects empty email",
			input: input{
				UserAdminId:         10,
				AdminPrivilegeGroup: "SuperAdmin",
				Username:            "newadmin2",
				Email:               "",
				Password:            "helloWorld2",
				FullName:            "John Smith2",
				PrivilegeGroup:      "Auditor",
			},
			expected: expected{
				success:          false,
				errorCode:        http.StatusBadRequest,
				assignCalls:      0,
				privilegeCalls:   0,
				groupExistsCalls: 0,
			},
			privilegeExistsMock:       defaultPrivilegeExistsMock(),
			hasPrivilegeMock:          defaultHasPrivilegeMock(),
			changeAdminPermissionMock: defaultChangeAdminPermissionMock(),
		},
		{
			name: "rejects empty password",
			input: input{
				UserAdminId:         10,
				AdminPrivilegeGroup: "SuperAdmin",
				Username:            "newadmin2",
				Email:               "new-admin2@gmail.com",
				Password:            "",
				FullName:            "John Smith2",
				PrivilegeGroup:      "Auditor",
			},
			expected: expected{
				success:          false,
				errorCode:        http.StatusBadRequest,
				assignCalls:      0,
				privilegeCalls:   0,
				groupExistsCalls: 0,
			},
			privilegeExistsMock:       defaultPrivilegeExistsMock(),
			hasPrivilegeMock:          defaultHasPrivilegeMock(),
			changeAdminPermissionMock: defaultChangeAdminPermissionMock(),
		},
		{
			name: "rejects empty full name",
			input: input{
				UserAdminId:         10,
				AdminPrivilegeGroup: "SuperAdmin",
				Username:            "newadmin2",
				Email:               "new-admin2@gmail.com",
				Password:            "asdasdasd",
				FullName:            "",
				PrivilegeGroup:      "Auditor",
			},
			expected: expected{
				success:          false,
				errorCode:        http.StatusBadRequest,
				assignCalls:      0,
				privilegeCalls:   0,
				groupExistsCalls: 0,
			},
			privilegeExistsMock:       defaultPrivilegeExistsMock(),
			hasPrivilegeMock:          defaultHasPrivilegeMock(),
			changeAdminPermissionMock: defaultChangeAdminPermissionMock(),
		},
		{
			name: "rejects empty privilege group",
			input: input{
				UserAdminId:         10,
				AdminPrivilegeGroup: "SuperAdmin",
				Username:            "newadmin2",
				Email:               "new-admin2@gmail.com",
				Password:            "asdasdasd",
				FullName:            "John Smith",
				PrivilegeGroup:      "",
			},
			expected: expected{
				success:          false,
				errorCode:        http.StatusBadRequest,
				assignCalls:      0,
				privilegeCalls:   0,
				groupExistsCalls: 0,
			},
			privilegeExistsMock:       defaultPrivilegeExistsMock(),
			hasPrivilegeMock:          defaultHasPrivilegeMock(),
			changeAdminPermissionMock: defaultChangeAdminPermissionMock(),
		},
		{
			name: "rejects a privilege group that does not exist",
			input: input{
				UserAdminId:         10,
				AdminPrivilegeGroup: "SuperAdmin",
				Username:            "new-admin2",
				Email:               "new-admin2@gmail.com",
				Password:            "helloWorld2",
				FullName:            "John Smith2",
				PrivilegeGroup:      "Auditor",
			},
			expected: expected{
				success:          false,
				errorCode:        http.StatusBadRequest,
				assignCalls:      0,
				groupExistsCalls: 1,
			},
			privilegeExistsMock: fixture.Modify(defaultPrivilegeExistsMock, func(out *permissionSvc.CheckAdminGroupPrivilegeExistsOut) {
				out.IsExists = false
			}),
			hasPrivilegeMock:          defaultHasPrivilegeMock(),
			changeAdminPermissionMock: defaultChangeAdminPermissionMock(),
		},
		{
			name: "check privilege exists failing",
			input: input{
				UserAdminId:         10,
				AdminPrivilegeGroup: "SuperAdmin",
				Username:            "new-admin2",
				Email:               "new-admin2@gmail.com",
				Password:            "helloWorld2",
				FullName:            "John Smith2",
				PrivilegeGroup:      "Auditor",
			},
			expected: expected{
				success:          false,
				errorCode:        http.StatusInternalServerError,
				assignCalls:      0,
				groupExistsCalls: 1,
			},
			privilegeExistsMock: fixture.Modify(defaultPrivilegeExistsMock, func(out *permissionSvc.CheckAdminGroupPrivilegeExistsOut) {
				out.Success = false
				out.ErrorCode = http.StatusInternalServerError
			}),
			hasPrivilegeMock:          defaultHasPrivilegeMock(),
			changeAdminPermissionMock: defaultChangeAdminPermissionMock(),
		},
		{
			name: "rejects caller without registration privilege",
			input: input{
				UserAdminId:         10,
				AdminPrivilegeGroup: "SuperAdmin",
				Username:            "new-admin2",
				Email:               "new-admin2@gmail.com",
				Password:            "helloWorld2",
				FullName:            "John Smith2",
				PrivilegeGroup:      "Auditor",
			},
			expected: expected{
				success:          false,
				errorCode:        http.StatusForbidden,
				privilegeCalls:   1,
				groupExistsCalls: 1,
			},
			privilegeExistsMock: defaultPrivilegeExistsMock(),
			hasPrivilegeMock: fixture.Modify(defaultHasPrivilegeMock, func(out *permissionSvc.HasPrivilegeOut) {
				out.IsAllowed = false
			}),
			changeAdminPermissionMock: defaultChangeAdminPermissionMock(),
		},
		{
			name: "fail when checking privilege failing",
			input: input{
				UserAdminId:         10,
				AdminPrivilegeGroup: "SuperAdmin",
				Username:            "new-admin2",
				Email:               "new-admin2@gmail.com",
				Password:            "helloWorld2",
				FullName:            "John Smith2",
				PrivilegeGroup:      "Auditor",
			},
			expected: expected{
				success:          false,
				errorCode:        http.StatusInternalServerError,
				privilegeCalls:   1,
				groupExistsCalls: 1,
			},
			privilegeExistsMock: defaultPrivilegeExistsMock(),
			hasPrivilegeMock: fixture.Modify(defaultHasPrivilegeMock, func(out *permissionSvc.HasPrivilegeOut) {
				out.Success = false
				out.ErrorCode = http.StatusInternalServerError
			}),
			changeAdminPermissionMock: defaultChangeAdminPermissionMock(),
		},
		{
			name: "fail when change permission failing",
			input: input{
				UserAdminId:         10,
				AdminPrivilegeGroup: "SuperAdmin",
				Username:            "new-admin2",
				Email:               "new-admin2@gmail.com",
				Password:            "helloWorld2",
				FullName:            "John Smith2",
				PrivilegeGroup:      "Auditor",
			},
			expected: expected{
				success:          false,
				errorCode:        http.StatusInternalServerError,
				assignCalls:      1,
				privilegeCalls:   1,
				groupExistsCalls: 1,
			},
			privilegeExistsMock: defaultPrivilegeExistsMock(),
			hasPrivilegeMock:    defaultHasPrivilegeMock(),
			changeAdminPermissionMock: fixture.Modify(defaultChangeAdminPermissionMock, func(out *permissionSvc.ChangeAdminPrivilegeRoleOut) {
				out.Success = false
				out.ErrorCode = http.StatusInternalServerError
			}),
		},
		{
			name: "rejects duplicate username",
			input: input{
				UserAdminId:         10,
				AdminPrivilegeGroup: "SuperAdmin",
				Username:            "taken",
				Email:               "new-admin2@gmail.com",
				Password:            "helloWorld2",
				FullName:            "John Smith2",
				PrivilegeGroup:      "Auditor",
			},
			expected: expected{
				success:          false,
				errorCode:        http.StatusBadRequest,
				assignCalls:      0,
				privilegeCalls:   1,
				groupExistsCalls: 1,
			},
			privilegeExistsMock:       defaultPrivilegeExistsMock(),
			hasPrivilegeMock:          defaultHasPrivilegeMock(),
			changeAdminPermissionMock: defaultChangeAdminPermissionMock(),
		},
		{
			name: "rejects duplicate email",
			input: input{
				UserAdminId:         10,
				AdminPrivilegeGroup: "SuperAdmin",
				Username:            "new-admin2",
				Email:               "taken@example.test",
				Password:            "helloWorld2",
				FullName:            "John Smith2",
				PrivilegeGroup:      "Auditor",
			},
			expected: expected{
				success:          false,
				errorCode:        http.StatusBadRequest,
				assignCalls:      0,
				privilegeCalls:   1,
				groupExistsCalls: 1,
			},
			privilegeExistsMock:       defaultPrivilegeExistsMock(),
			hasPrivilegeMock:          defaultHasPrivilegeMock(),
			changeAdminPermissionMock: defaultChangeAdminPermissionMock(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture.NewUserFixture(t)
			f.SeedUsers(t, []fixture.UserRow{
				{Username: "taken", Email: "taken@example.test"},
			})

			f.PermissionService.CheckAdminGroupPrivilegeExistsFunc = func(ctx context.Context, in *permissionSvc.CheckAdminGroupPrivilegeExistsIn) *permissionSvc.CheckAdminGroupPrivilegeExistsOut {
				testassert.AssertEqual(t, "CheckAdminGroupPrivilegeExists.AdminPrivilegeGroup", tt.input.PrivilegeGroup, in.AdminPrivilegeGroup)
				return tt.privilegeExistsMock
			}
			f.PermissionService.HasPrivilegeFunc = func(ctx context.Context, in *permissionSvc.HasPrivilegeIn) *permissionSvc.HasPrivilegeOut {
				testassert.AssertEqual(t, "HasPrivilege.PrivilegeGroup", tt.input.AdminPrivilegeGroup, in.PrivilegeGroup)
				testassert.AssertEqual(t, "HasPrivilege.Privilege", permission.PVRegisterAdminUser, in.Privilege)
				return tt.hasPrivilegeMock
			}
			f.PermissionService.ChangeAdminPrivilegeRoleFunc = func(ctx context.Context, in *permissionSvc.ChangeAdminPrivilegeRoleIn) *permissionSvc.ChangeAdminPrivilegeRoleOut {
				testassert.AssertEqual(t, "ChangeAdminPrivilegeRole.AdminPrivilegeGroup", tt.input.AdminPrivilegeGroup, in.AdminPrivilegeGroup)
				testassert.AssertEqual(t, "ChangeAdminPrivilegeRole.NewPrivilegeGroup", tt.input.PrivilegeGroup, in.NewPrivilegeGroup)
				testassert.AssertEqual(t, "ChangeAdminPrivilegeRole.UserAdminId", tt.input.UserAdminId, in.UserAdminId)

				// Since target user id is newly created we just have to ensure that it's not zero.
				testassert.AssertNotZero(t, "ChangeAdminPrivilegeRole.TargetUserID", in.TargetUserID)
				return tt.changeAdminPermissionMock
			}

			initialUserCount := f.UserQuery.CountUsers()

			startTs, err := helpers.TimeNow()
			if err != nil {
				t.Fatal(err)
			}

			got := f.AdminService.RegisterNewAdmin(context.Background(),
				&userSvc.RegisterNewAdminIn{
					Trace:               fixture.TestTrace(),
					UserAdminId:         tt.input.UserAdminId,
					AdminPrivilegeGroup: tt.input.AdminPrivilegeGroup,
					Username:            tt.input.Username,
					Email:               tt.input.Email,
					Password:            tt.input.Password,
					FullName:            tt.input.FullName,
					PrivilegeGroup:      tt.input.PrivilegeGroup,
				})
			testassert.AssertEqual(t, "RegisterNewAdmin() success", tt.expected.success, got.Success)
			testassert.AssertEqual(t, "RegisterNewAdmin() error code", tt.expected.errorCode, got.ErrorCode)

			if calls := f.PermissionService.ChangeAdminPrivilegeRoleCalls(); len(calls) != tt.expected.assignCalls {
				t.Fatalf("ChangeAdminPrivilegeRole calls = %d, want %d", len(calls), tt.expected.assignCalls)
			}
			if calls := f.PermissionService.HasPrivilegeCalls(); len(calls) != tt.expected.privilegeCalls {
				t.Fatalf("HasPrivilege calls = %d, want %d", len(calls), tt.expected.privilegeCalls)
			}
			if calls := f.PermissionService.CheckAdminGroupPrivilegeExistsCalls(); len(calls) != tt.expected.groupExistsCalls {
				t.Fatalf("CheckAdminGroupPrivilegeExists calls = %d, want %d", len(calls), tt.expected.groupExistsCalls)
			}

			count := f.UserQuery.CountUsers()
			if tt.expected.success == false {
				// User not changed.
				testassert.AssertEqual(t, "UserCount", initialUserCount, count)
				return
			}

			insertedUser := f.UserQuery.FindByID(got.InsertedUserId)
			testassert.AssertNil(t, "RegisterNewAdmin inserted user DeletedAt", insertedUser.DeletedAt)
			testassert.AssertNotEmpty(t, "RegisterNewAdmin inserted user ExternalID", insertedUser.ExternalID)
			testassert.AssertEqual(t, "RegisterNewAdmin inserted user Username", tt.input.Username, insertedUser.Username)
			testassert.AssertEqual(t, "RegisterNewAdmin inserted user Email", tt.input.Email, insertedUser.Email)
			testassert.AssertTrue(t, "RegisterNewAdmin check inserted user", helpers.VerifyPassword(tt.input.Password, insertedUser.Password))
			testassert.AssertEqual(t, "RegisterNewAdmin inserted user FullName", tt.input.FullName, insertedUser.FullName)
			testassert.AssertEqual(t, "RegisterNewAdmin inserted user Role", "Admin", insertedUser.Role)
			testassert.AssertTimeAfter(t, "RegisterNewAdmin inserted user CreatedTs", startTs, insertedUser.Timestamp.CreatedTs)
			testassert.AssertTimeAfter(t, "RegisterNewAdmin inserted user LastUpdatedTs", startTs, insertedUser.Timestamp.LastUpdatedTs)
			testassert.AssertEqual(t, "UserCount", initialUserCount+1, count)
		})
	}
}
