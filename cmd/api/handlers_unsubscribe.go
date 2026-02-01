/** ****************************************************************************************************************** **
	for webhooks from email providers
** ****************************************************************************************************************** **/

package main 

import (
	"coldbrew/tools"
	"coldbrew/db/postgres"

	"github.com/gofiber/fiber/v2"
	
	"net/http"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- TEAMS -------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

// renders the webpage for a user to unsubscribe from

func (this *app) unsubscribeGet (c *fiber.Ctx) error {
	ctx, cancel := handlerCtx()
	defer cancel()

	var token tools.String 
	token.Set (c.Params("token"))

	// make sure we have this token in our database, else return a 404
	user, err := this.db.UserFromBearer (ctx, token)
	if err != nil {
		return this.Respond (ctx, err, c, nil)
	}

	if user == nil {
		return this.RespondError (ctx, nil, c, http.StatusNotFound, "") // not a user
	}

	return c.Render ("unsubscribe", fiber.Map {
		"ServiceName": cfg.ServiceName,
		"BaseUrl": cfg.ApiUrl,
		"UserToken": token.String(),
	})
}

func (this *app) unsubscribePut (c *fiber.Ctx) error {
	ctx, cancel := handlerCtx()
	defer cancel()

	var token tools.String 
	token.Set (c.Params("token"))

	// make sure we have this token in our database, else return a 404
	user, err := this.db.UserFromBearer (ctx, token)
	if err != nil {
		return this.Respond (ctx, err, c, nil)
	}

	if user == nil {
		return this.RespondError (ctx, nil, c, http.StatusNotFound, "") // not a user
	}

	err = this.db.UserSetMask (ctx, user, postgres.UserMask_unsubscribe)

	return this.Respond (ctx, err, c, nil)
}
