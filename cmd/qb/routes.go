/** ****************************************************************************************************************** **
	Endpoints supported by the API
	
** ****************************************************************************************************************** **/

package main

import (
	"coldbrew/tools"

	"github.com/gofiber/fiber/v2"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- LOCAL MIDDLEWARE --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *app) defaultGet (c *fiber.Ctx) error {
	return c.SendString(serviceName + " " + serviceVersion)
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- ROUTES ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *app) ready (c *fiber.Ctx) error {
	if this.Running == false {
		return this.K8ServiceNotRunning (c)
	}

	ctx, cancel := tools.TimeDuration(2).Context ("health check ready")
	defer cancel()

	this.db.Ping(ctx) // this doesn't appear to error, and I know it's supposed to try the connect again auto-magically, so just call it each time

	if err := this.db.HealthCheck (ctx); err != nil {
		this.StackTrace(ctx, err)
		return this.K8ServiceNotRunning (c)
	}

	return c.SendString("We're good") 
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- HANDLERS ----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *app) routes () *fiber.App {
	// standard chain that all calls make
	app := this.Routes()

	app.Get("/", this.defaultGet)

	app.Get("/status/ready", this.ready)

	return app
}
