/** ****************************************************************************************************************** **
	Flow logic for asking AI about the hypothetical matchups between a player and a team
	
** ****************************************************************************************************************** **/

package main

import (
	"coldbrew/db/postgres"
	"coldbrew/tools/zerobounce"
	
	"context"
	"log/slog"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// i'm doing it this way to try to reduce the number of functions in our app class
// i think it helps keep the code a little cleaner without having to do the hassle of moving this
// flow stuff to yet another folder
type flowEmailValidate struct {
	*app
}

// validates the user's email
func (this *flowEmailValidate) validateUser (ctx context.Context, user *postgres.User) error {

	// make sure this email appears valid before we even bother with zerobounce
	if user.Email.Email() == false {
		return this.db.UserSetDisabled (ctx, user) // disable them, they're not good
	}

	// validate this with our bounce config
	if cfg.ZeroBounce.Valid() {
		valid, typo, err := zerobounce.ValidateEmail (ctx, cfg.ZeroBounce.String(), user.Email.String())
		if err != nil { return err }

		if valid == false && len(typo) > 0 {
			// we think there was a typo, so let's just update this user's email
			user.Email.Set(typo)
			if err := this.db.UserUpdate (ctx, user); err != nil { return err }

			valid = true // let's just say they're valid now
		}

		if valid {
			return this.db.UserSetValid (ctx, user) // they're good
		} else {
			return this.db.UserSetDisabled (ctx, user) // disable them, they're not good
		}
		
	}
	
	slog.Info ("no zero bounce api key found, all emails will be sent to")
	return this.db.UserSetValid (ctx, user) // just say they're good
}

// finds the next combination of team/player that we need to ask ai about
func (this *app) flowEmailValidate (ctx context.Context) error {
	// generate our class
	flow := &flowEmailValidate { app: this } // share the pointer

	user, err := this.db.UserNotValidated (ctx)
	if err != nil { return err }

	if user != nil {
		err = flow.validateUser (ctx, user)
	}
	return err // just the one
}
