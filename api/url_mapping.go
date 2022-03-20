package api

import (
	"github.com/gofiber/fiber/v2"
	communityApis "github.com/kotalco/api/api"
	"github.com/kotalco/cloud-api/api/handler/user"
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
	users.Get("/totp", middleware.JWTProtected, user.CreateTOTP)
	users.Post("/totp/enable", middleware.JWTProtected, user.EnableTwoFactorAuth)
	users.Post("/totp/verify", middleware.JWTProtected, user.VerifyTOTP)
	users.Post("/totp/disable", middleware.JWTProtected, middleware.TFAProtected, user.DisableTwoFactorAuth)

	//community routes
	communityApis.MapUrl(app, middleware.JWTProtected, middleware.TFAProtected)
}
