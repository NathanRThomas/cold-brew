/** ****************************************************************************************************************** **
	Flow logic for asking AI about the hypothetical matchups between a player and a team
	
** ****************************************************************************************************************** **/

package main

import (
	"coldbrew/db/postgres"
	
	"github.com/pkg/errors"

	"context"
	"time"
	"math/rand"
	"log/slog"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// i'm doing it this way to try to reduce the number of functions in our app class
// i think it helps keep the code a little cleaner without having to do the hassle of moving this
// flow stuff to yet another folder
type flowQueEmails struct {
	*app
}

// for when we want to schedule text based emails to start things off
func (this *flowQueEmails) textWarmup (ctx context.Context, mailman *postgres.Mailman) error {
	// get our warmup users
	users, err := this.db.UsersFromMask (ctx, postgres.UserMask_warmup)
	if err != nil { return err }

	if len(users) < 5 { return errors.Errorf("no enough warmup users. use at least 5") }

	// for this we que all of them up, 1 hour apart
	// let's find our warmup template
	templates, err := this.db.TemplateList (ctx)
	if err != nil { return err }
	if len(templates) == 0 { return errors.Errorf("No tempaltes found for initial text warmup") }

	target := templates[0] // just default to the first

	// look for the warmup one
	for _, template := range templates {
		if template.Mask & postgres.TemplateMask_warmup > 0 {
			target = template
			break
		}
	}

	// before we schedule the emails, let's make sure we can update the mask
	if mailman.Mask & postgres.MailmanMask_textWarm == 0 {
		if err := this.db.MailmanSetMask (ctx, mailman, postgres.MailmanMask_textWarm); err != nil { return err }
	} else {
		// set the html version now
		if err := this.db.MailmanSetMask (ctx, mailman, postgres.MailmanMask_htmlWarm); err != nil { return err }
	}

	// que up the warmup emails
	targetToQue := 30
	nextEmail := time.Now() // start right away for this one
	for targetToQue > 0 {
		for _, user := range users {
			targetToQue--

			email := &postgres.Email {
				Mailman: mailman.Id,
				Template: target.Id,
				User: user.Id,
				Target: nextEmail,
			}

			// now insert it
			if err := this.db.EmailInsert (ctx, email); err != nil { return err }

			// these are once an hour or so, i think we want jitter in there
			nextEmail = nextEmail.Add(time.Minute * time.Duration(50 + rand.Intn(10)))
		}
	}

	// we're good, this mailman has their plain text warmups set
	return nil
}

// checks a mailmen to see what we need to do next with them
func (this *flowQueEmails) mailman (ctx context.Context, mailman *postgres.Mailman) error {

	// if this already has any scheduled emails, then we're done
	cnt, err := this.db.EmailsScheduledByMailman (ctx, mailman.Id)
	if err != nil || cnt > 0 { return err } // we're done either way

	// no future scheduled emails, see what the status is
	if mailman.Mask & (postgres.MailmanMask_textWarm | postgres.MailmanMask_htmlWarm) != (postgres.MailmanMask_textWarm | postgres.MailmanMask_htmlWarm) {
		// we're still warming this one up
		return this.textWarmup (ctx, mailman)
	}

	// we're warmed up, so let's find some non-warmup templates to send
	templates, err := this.db.TemplateList (ctx)
	if err != nil { return err }
	
	targetTemplates := make([]*postgres.Template, 0, len(templates))

	for _, template := range templates {
		if template.Mask & postgres.TemplateMask_warmup == 0 {
			targetTemplates = append (targetTemplates, template) // this one counts
		}
	}

	if len(targetTemplates) == 0 { return errors.Errorf("No templates found for mailman : %s", mailman.Id) }

	// we want a list of the users that haven't gotten an email yet
	users, err := this.db.UsersMissing (ctx)
	if err != nil { return err }

	if len(users) == 0 {
		slog.Warn("no more users to send to")
		return nil 
	}

	nextEmail := time.Now() // send the next one right away
	// get our performance for this mailman
	frequency, err := this.db.MailmanPerformance (ctx, mailman.Id)
	if err != nil { return err }

	templateIdx := 0 // start with the first template and a/b from there

	// loop through all the users we pulled in
	for _, user := range users {
		email := &postgres.Email {
			Mailman: mailman.Id,
			Template: targetTemplates[templateIdx].Id,
			User: user.Id,
			Target: nextEmail,
		}

		// now insert it
		if err := this.db.EmailInsert (ctx, email); err != nil { return err }

		nextEmail = nextEmail.Add(frequency)
		// i think we want jitter in there
		nextEmail = nextEmail.Add(time.Second * time.Duration(rand.Intn(20)))

		templateIdx++
		if templateIdx >= len(targetTemplates) { templateIdx = 0 } // reset this
	}

	return nil // we're good
}

// finds the next combination of team/player that we need to ask ai about
func (this *app) flowQueEmails (ctx context.Context) error {
	// generate our class
	flow := &flowQueEmails { app: this } // share the pointer

	// start with the mailmen
	mailmen, err := this.db.MailmanList (ctx)
	if err != nil { return err }

	for _, mailman := range mailmen {
		this.StackTrace (ctx, flow.mailman (ctx, mailman))
	}
	return nil // we're done
}
