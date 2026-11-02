package application

type Service struct {
	*RegisterUseCase
	*LoginUseCase
	*RefreshTokenUseCase
	*LogoutUseCase
	*GetSelfProfileUseCase
	*GetUserByIDUseCase
	*AddRoleUseCase
	*RemoveRoleUseCase
	*GetPublicProfileUseCase
	*UpdateProfileUseCase
	*ChangePasswordUseCase
	*AdminUpdateProfileUseCase
	*AdminResetPasswordUseCase
	*ListUsersUseCase
	*GetUserRolesUseCase
}

func NewService(
	register *RegisterUseCase,
	login *LoginUseCase,
	refreshToken *RefreshTokenUseCase,
	logout *LogoutUseCase,
	getSelfProfile *GetSelfProfileUseCase,
	getUserByID *GetUserByIDUseCase,
	addRole *AddRoleUseCase,
	removeRole *RemoveRoleUseCase,
	getPublicProfile *GetPublicProfileUseCase,
	updateProfile *UpdateProfileUseCase,
	changePassword *ChangePasswordUseCase,
	adminUpdateProfile *AdminUpdateProfileUseCase,
	adminResetPassword *AdminResetPasswordUseCase,
	listUsersUseCase *ListUsersUseCase,
	getUserRolesUseCase *GetUserRolesUseCase,
) *Service {
	return &Service{
		RegisterUseCase:           register,
		LoginUseCase:              login,
		RefreshTokenUseCase:       refreshToken,
		LogoutUseCase:             logout,
		GetSelfProfileUseCase:     getSelfProfile,
		GetUserByIDUseCase:        getUserByID,
		AddRoleUseCase:            addRole,
		RemoveRoleUseCase:         removeRole,
		GetPublicProfileUseCase:   getPublicProfile,
		UpdateProfileUseCase:      updateProfile,
		ChangePasswordUseCase:     changePassword,
		AdminUpdateProfileUseCase: adminUpdateProfile,
		AdminResetPasswordUseCase: adminResetPassword,
		ListUsersUseCase:          listUsersUseCase,
		GetUserRolesUseCase:       getUserRolesUseCase,
	}
}
