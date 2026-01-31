/** ****************************************************************************************************************** **
	Reused functions by the handlers
	
** ****************************************************************************************************************** **/

package main

import (
	"coldbrew/tools"
	"coldbrew/tools/logging"
	
	"github.com/google/uuid"
	"github.com/gofiber/fiber/v2"
	
	"context"
	"net/http"
	"strings"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

var (
	handlerTimeout 	tools.TimeDuration = 29 // seconds for a request to finish before we bail on it
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- HELPERS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func handlerCtx () (context.Context, context.CancelFunc) {
	ctx, cancel := handlerTimeout.Context("api")

	u := uuid.New()
	ctx = context.WithValue (ctx, logging.ThreadId, u.String())

	return ctx, cancel
}

func getVersion (c *fiber.Ctx) int {
	var version tools.String
	version.Set(c.Get ("api-version"))
	v := version.Int()
	if v <= 1 { return 1 }

	return v // whatever version we're on
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- BEARER ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

// for when we make actions makes calls to us
func (this *app) bearer (c *fiber.Ctx) error {
	ctx, cancel := handlerCtx()
	defer cancel()

	auth := c.Get ("Authorization")

	// we have a bearer to care about
	bearer := strings.Split (auth, "Bearer")
	if len(bearer) == 2 {
		auth = strings.TrimSpace (bearer[1]) // re-use this as the auth
	}

	if len(auth) == 0 {
		// they didn't pass a bearer
		return this.RespondError (ctx, nil, c, http.StatusUnauthorized, "Bearer appears invalid")
	}

	if len(auth) > 1000 { // thinking about this, could be exploited by a buffer overrun, no reason for the auth to be this large
		return this.RespondError (ctx, nil, c, http.StatusUnauthorized, "Bearer appears invalid")
	}

	if auth == cfg.AdminToken {
		// this is good, it's our key for github secrets
	} else {
		return this.RespondError (ctx, nil, c, http.StatusUnauthorized, "Bearer appears invalid")
	}

	return c.Next() // continue along
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PATH PARAMS -------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//
