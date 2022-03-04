package api

import (
	"github.com/gofiber/fiber/v2"
	apiCommunity "github.com/kotalco/api/api"
)

// MapUrl abstracted function to map and register all the url for the application
func MapUrl(app *fiber.App) {

	apiCommunity.MapUrl(app)


}
