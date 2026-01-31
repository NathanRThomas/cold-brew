/** ****************************************************************************************************************** **
	Endpoints related to games
** ****************************************************************************************************************** **/

package main 

import (
	"coldbrew/tools"
	"coldbrew/tools/logging"
	
	"github.com/pkg/errors"
	"github.com/gofiber/fiber/v2"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type userPutRequest struct {
	Warmup bool
	Emails tools.StringList
}

// validates the data is ok to create
func (this *userPutRequest) ValidInput () error {
	if this.Emails.Len() == 0 {
		return errors.Wrap (logging.ErrReturnToUser, "no emails found")
	}

	return nil // we're good
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- USERS -------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

// ads more user emails
func (this *app) userPut (c *fiber.Ctx) error {
	ctx, cancel := handlerCtx()
	defer cancel()
	
	// user := c.Locals(userCtxKey).(*postgres.User)

	data := &userPutRequest{}
	if this.ValidateInput (ctx,c, data) == false {
		return nil
	}

	// resp, err := this.api.UpdateUser (ctx, user, data.FirstName, data.LastName, data.Timezone)

	return this.Respond (ctx, nil, c, nil)
}
