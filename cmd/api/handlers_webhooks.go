/** ****************************************************************************************************************** **
	for webhooks from email providers
** ****************************************************************************************************************** **/

package main 

import (
	"coldbrew/tools"
	"coldbrew/db/postgres"

	"github.com/gofiber/fiber/v2"
	json "github.com/json-iterator/go"

	"fmt"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type sendgridPost []struct {
	Category tools.StringList
	Email, Event tools.String
	Sg_event_id, Sg_message_id tools.String
}

// validates the data is ok to create
func (this sendgridPost) ValidInput () error {
	return nil // we're good
}


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- TEAMS -------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//


func (this *app) sendgridPost (c *fiber.Ctx) error {
	ctx, cancel := handlerCtx()
	defer cancel()

	data := sendgridPost{}
	
	// record the body so we can debug
	bodyBytes := c.Body()
	fmt.Println("sendgridPost", string(bodyBytes))

	err := json.Unmarshal (bodyBytes, &data)
	if err != nil {
		this.StackTrace (ctx, err)
		// this is bad
	} else {
		// this worked, we got some data

		for _, event := range data {
			// for each piece of data we have some possible action items
			// first let's give a report about this user
			if event.Email.Email() {
				this.StackTrace (ctx, this.db.UserUpdateStatus (ctx, event.Email, event.Event.String()))
			} else {
				this.TraceErr (ctx, "we got an invalid email: %s", string(bodyBytes))
			}

			// now let's give a report about this email we sent
			if event.Sg_message_id.Valid() {
				this.StackTrace (ctx, this.db.EmailUpdateStatus (ctx, event.Email, event.Sg_message_id.String(), postgres.EmailStatus(event.Event.String())))
			} else {
				this.TraceErr (ctx, "we got an invalid sg_message_id: %s", string(bodyBytes))
			}
		}
	}

	return this.LiveCheck (c)
}
