/** ****************************************************************************************************************** **
	Flow logic for sending emails. right now it's just one at a time and database heavy
	Could use caching
	
** ****************************************************************************************************************** **/

package main

import (
	"coldbrew/db/postgres"
	"coldbrew/tools/sendgrid"

	"context"
	"time"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// i'm doing it this way to try to reduce the number of functions in our app class
// i think it helps keep the code a little cleaner without having to do the hassle of moving this
// flow stuff to yet another folder
type flowEmailSend struct {
	*app
}

// looks in the database for emails that need to be sent
func (this *flowEmailSend) emails (ctx context.Context) error {
	// find the next email
	email, err := this.db.EmailsToSend (ctx)
	if err != nil { return err }
	if email == nil { 
		time.Sleep(time.Second * 3) // give a little break as there was nothing to do last time
		return nil // we're good, just nothing needs to be sent
	}
	
	// we need to get all this data about the email
	user, err := this.db.User (ctx, email.User)
	if err != nil { return err }

	// now the template
	template, err := this.db.Template (ctx, email.Template)
	if err != nil { return err }

	// now the mailman for sending
	mailman, err := this.db.Mailman (ctx, email.Mailman)
	if err != nil { return err }

	// we have what we need, let's generate the text
	htmlBody := ""
	textBody, err := template.GenerateTextBody()
	if err != nil { return err }

	if mailman.Mask & postgres.MailmanMask_htmlWarm > 0 { // only include this if we're warmed up
		body, err := template.GenerateHTMLBody()
		if err != nil { return err }
		htmlBody = body // copy this over
	}

	// record this as sent in the database, so we don't keep sending the user emails
	if err := this.db.EmailSent (ctx, email.Id); err != nil { return err }

	// we're finally ready to send this
	go func() { // this creates its own context, so just go with that
		err := sendgrid.SendEmail (mailman.Attr.APIToken.String(), user.Email.String(), mailman.Attr.Category.String(),
			template.GenerateSubject(), textBody, htmlBody, mailman.Attr.IpPool.String(), mailman.Attr.FromEmail.String(),
			mailman.Attr.FromName.String(), mailman.Attr.ReplyName.String(), mailman.Attr.ReplyEmail.String())
		if err != nil {
			this.StackTrace (ctx, err) // record this
		}
	}()

	return nil // we're good
}

// Primary call
// this will check the database for emails that need to be sent
func (this *app) flowEmailSend (ctx context.Context) error {

	// generate our class
	flow := &flowEmailSend { app: this } // share the pointer

	return flow.emails (ctx)
}
