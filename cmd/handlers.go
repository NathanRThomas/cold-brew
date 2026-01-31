/** ****************************************************************************************************************** **
	Re-usable functions used by the api handlers

** ****************************************************************************************************************** **/

package cmd

import (
	"coldbrew/tools"
	"coldbrew/db"
	"coldbrew/tools/logging"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	
	"net/http"
	"context"
	"math/rand"
	"time"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- INTERFACES --------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type InputStruct interface {
	ValidInput () error
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS --------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type ErrReturn struct {
	Msg string `json:"msg"`
}


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (this *App) randError () string {
	terms := [...]string {"There was an issue with the monkies", "Houston we have a problem", "The gerbil stopped running",
		"Looks like the Fatherboard is burned out", "There's a loose wire between the mouse and keyboard", 
		"The monkies got out of their cages", "I think there's a gas leak", "Russia is hacking us", 
		"North Korea is attacking", "China is hacking us", "You can't get there from here" }

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := rnd.Intn(len(terms) - 1)
	return terms[n]
}

func (this *App) serverError (ctx context.Context, err error, c *fiber.Ctx) error {
	return this.RespondError (ctx, err, c, http.StatusInternalServerError, this.randError())
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

// ClientIP returns the real client IP from the X-Forwarded-For chain.
// When behind a load balancer, c.IPs() returns [client, proxy1, proxy2, ...].
// This returns the first IP (the actual client) or falls back to c.IP().
func (this *App) ClientIP (c *fiber.Ctx) string {
	ips := c.IPs()
	if len(ips) > 0 {
		return ips[0]
	}
	return c.IP() // fallback for direct connections
}

// Validates the input by the user to an endpoint
func (this *App) ValidateInput (ctx context.Context, c *fiber.Ctx, out InputStruct) bool {
	if err := c.BodyParser(out); err != nil {
		this.RespondError (ctx, errors.WithStack(err), c, http.StatusBadRequest, "json appears invalid")
		return false
	}
	
	if err := out.ValidInput(); err != nil {
		this.RespondError (ctx, nil, c, http.StatusBadRequest, err.Error())
		return false
	}
	return true // we're good
}

func (this *App) RespondWithStatus (ctx context.Context, err error, c *fiber.Ctx, httpStatus int, out interface{}) error {
	if err != nil {
		return this.Respond (ctx, err, c, out) // handle the error instead
	}

	if httpStatus <= http.StatusOK {
		httpStatus = http.StatusOK // if it's not set, but we have no error, then we're good
	}

	if out == nil { 
		var ret struct{}
		return c.Status(httpStatus).JSON(ret) // always return json
	}

	return c.Status(httpStatus).JSON(out)
}

func (this *App) RespondError (ctx context.Context, err error, c *fiber.Ctx, httpStatus int, msg string) error {
	if err != nil { this.StackTrace (ctx, err) }  // record this

	if len(msg) == 0 { msg = http.StatusText(httpStatus) }	// default to the text version of the status code

	return c.Status(httpStatus).JSON(ErrReturn { msg })
}

func (this *App) Respond (ctx context.Context, err error, c *fiber.Ctx, out interface{}) error {
	switch errors.Cause (err) {
	case logging.ErrReturnToUser:
		var msg tools.String
		msg.Set (err.Error())

		return this.RespondError (ctx, nil, c, http.StatusBadRequest, msg.Remove (":? ?Problem with the request"))

	case logging.ErrFunds:
		var msg tools.String
		msg.Set (err.Error())

		return this.RespondError (ctx, nil, c, http.StatusPaymentRequired, msg.Remove (":? ?You can't afford this"))

	case logging.ErrKey:
		return this.RespondError (ctx, nil, c, http.StatusLocked, err.Error())

	case logging.ErrPhone:
		return this.RespondError (ctx, nil, c, http.StatusNotAcceptable, err.Error())

	case logging.ErrUnauthorized:
		return this.RespondError (ctx, nil, c, http.StatusUnauthorized, err.Error())

	case nil:
		return this.RespondWithStatus (ctx, nil, c, http.StatusOK, out)
		
	case db.ErrKeyNotFound:
		return c.SendStatus(http.StatusNotFound)

	default:
		return this.serverError (ctx, err, c) // another error happened
	}
}
