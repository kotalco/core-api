package api

import (
	"github.com/gofiber/fiber/v2"
	communityApis "github.com/kotalco/api/api"
	"github.com/kotalco/cloud-api/api/handler/user"
	"github.com/kotalco/cloud-api/api/handler/workspace"
	"github.com/kotalco/cloud-api/pkg/middleware"
)

// MapUrl abstracted function to map and register all the url for the application
func MapUrl(app *fiber.App) {
	api := app.Group("api")
	v1 := api.Group("v1")
	//users group
	v1.Post("sessions", user.SignIn)
	users := v1.Group("users")
	users.Post("/", user.SignUp)
	users.Post("/resend_email_verification", user.SendEmailVerification)
	users.Post("/forget_password", user.ForgetPassword)
	users.Post("/reset_password", user.ResetPassword)
	users.Post("/verify_email", user.VerifyEmail)

	users.Post("/change_password", middleware.JWTProtected, middleware.TFAProtected, user.ChangePassword)
	users.Post("/change_email", middleware.JWTProtected, middleware.TFAProtected, user.ChangeEmail)
	users.Get("/whoami", middleware.JWTProtected, middleware.TFAProtected, user.Whoami)

	users.Post("/totp", middleware.JWTProtected, user.CreateTOTP)
	users.Post("/totp/enable", middleware.JWTProtected, user.EnableTwoFactorAuth)
	users.Post("/totp/verify", middleware.JWTProtected, user.VerifyTOTP)
	users.Post("/totp/disable", middleware.JWTProtected, middleware.TFAProtected, user.DisableTwoFactorAuth)

	//workspace group
	workspaces := v1.Group("workspaces")
	workspaces.Use(middleware.JWTProtected, middleware.TFAProtected)
	workspaces.Post("/", workspace.Create)
	workspaces.Patch("/:id", middleware.IsWorkspace, workspace.Update)
	workspaces.Delete("/:id", middleware.IsWorkspace, workspace.Delete)
	workspaces.Get("/", workspace.GetByUserId)
	workspaces.Post("/:id/members", middleware.IsWorkspace, workspace.AddMember)
	workspaces.Post("/:id/leave", middleware.IsWorkspace, workspace.Leave)
	workspaces.Delete("/:id/members/:user_id", middleware.IsWorkspace, workspace.RemoveMember)
	workspaces.Get("/:id/members", middleware.IsWorkspace, workspace.Members)

	//community routes
	communityApis.MapUrl(app, middleware.JWTProtected, middleware.TFAProtected)
}
