/** ****************************************************************************************************************** **
	Re-usable functions used by the api routing functions
	
** ****************************************************************************************************************** **/

package cmd 

import (
	json "github.com/json-iterator/go"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/compress"
	
	"net/http"
	"time"
	"os"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- MIDDLEWARE --------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- ROUTES ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *App) LiveCheck (c *fiber.Ctx) error {
	// this gets called first, and then kubernetes waits for it to be "live"
	// so if the code is running at all, we want to serve the next request and return a 200 here
	// if this code isn't running, that's a signal to k8 the pod can be deleted
	return c.SendString("We're good")
}

func (this *App) K8ServiceNotRunning (c *fiber.Ctx) error {
	return c.Status(http.StatusServiceUnavailable).SendString("We're NOT good")
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- ENTRY POINTS ------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *App) Routes () *fiber.App {
	// init our app with some defaults
	app := fiber.New (fiber.Config {
		ReadTimeout: time.Minute,
		WriteTimeout: time.Minute,
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return this.Respond (nil, err, c, nil)
		},
		Views: html.New(os.Getenv ("TEMPLATES") + "/website", ".html"),
		// Trust proxy headers from Google Cloud Load Balancer
		// This makes c.IP() return the real client IP from X-Forwarded-For
		ProxyHeader: "X-Forwarded-For",
	})

	// Enable Gzip compression
    app.Use(compress.New(compress.Config{
        Level: compress.LevelBestSpeed, // Choose compression level
    }))

	// cors
	app.Use(cors.New())
	
	// Attach the recover middleware to catch panics
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true, // Optional: logs stack traces
	}))

	app.Get("/status/live", this.LiveCheck)

	return app
}
