/** ****************************************************************************************************************** **
	SQL queries related to the emails table
	
** ****************************************************************************************************************** **/

package postgres

import (
	"coldbrew/tools"
	"coldbrew/db"
	
	"github.com/google/uuid"
	"github.com/pkg/errors"
	
	"time"
	"context"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type EmailMask int64
const (
	
)

type EmailStatus string 
const (
	EmailStatus_processed	= EmailStatus("processed")
	EmailStatus_delivered	= EmailStatus("delivered")
	EmailStatus_deferred	= EmailStatus("deferred")
	EmailStatus_dropped		= EmailStatus("dropped")
	EmailStatus_bounce		= EmailStatus("bounce")
	EmailStatus_blocked		= EmailStatus("blocked")
	EmailStatus_open		= EmailStatus("open")
	EmailStatus_click		= EmailStatus("click")
	EmailStatus_spamreport	= EmailStatus("spamreport")
	EmailStatus_unsubscribe	= EmailStatus("unsubscribe")
	EmailStatus_groupUnsub	= EmailStatus("group_unsubscribe")
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type Email struct {
	db.DBStruct
	Mailman, Template, User *uuid.UUID
	Target time.Time
	MessageId tools.String
	Status EmailStatus
}

func (this *Email) Key () string {
	return this.PrefixKey("email")
}

func (this *Email) CacheTime () tools.TimeDuration {
	return tools.TimeDuration(60) //
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

func (this *Coldbrew) emailSetStatus (ctx context.Context, email *Email, status EmailStatus) error {
	if email.Status == status { return nil } // we're good

	email.Status = status // update in real time
	return this.Exec (ctx, nil, `UPDATE emails SET status = $2 WHERE id = $1`, email.Id, status)
}

// lists all the non-paused emails
func (this *Coldbrew) EmailInsert (ctx context.Context, email *Email) error {
	email.SetPK()

	return this.Exec (ctx, nil, `INSERT INTO emails (id, target_time, mailman, template, "user", status, message_id) 
									VALUES ($1, $2, $3, $4, $5, $6, $7)`, email.Id, email.Target, 
									email.Mailman, email.Template, email.User, email.Status, email.MessageId)
}

// quick check for future scheduled emails for a mailman
func (this *Coldbrew) EmailsScheduledByMailman (ctx context.Context, mailmanId *uuid.UUID) (cnt int, err error) {
	err = this.DB.QueryRow (ctx, `SELECT COUNT(*) FROM emails WHERE sent_time IS NULL AND mailman = $1`, mailmanId).Scan(&cnt)
	return
}

// finds the next email that needs to be sent
func (this *Coldbrew) EmailsToSend (ctx context.Context) (*Email, error) {
	email := &Email{}
	err := this.DB.QueryRow (ctx, `SELECT id, mailman, template, "user" FROM emails 
										WHERE sent_time IS NULL AND target_time < now() ORDER BY target_time LIMIT 1`).Scan(&email.Id,
										&email.Mailman, &email.Template, &email.User)
	if this.ErrNoRows(err) { return nil, nil } // nothing to send

	return email, errors.WithStack(err)
}

// marks the email as being sent
func (this *Coldbrew) EmailSent (ctx context.Context, emailId *uuid.UUID) error {
	return this.Exec (ctx, nil, `UPDATE emails SET sent_time = NOW() WHERE id = $1`, emailId)
}

func (this *Coldbrew) EmailUpdateStatus (ctx context.Context, userEmail tools.String, messageId string, status EmailStatus) error {
	email := &Email{}
	err := this.DB.QueryRow (ctx, `SELECT id, status FROM emails WHERE message_id = $1`, 
								messageId).Scan(&email.Id, &email.Status)
	if this.ErrNoRows (err) { 
		// this is ok, and expected as we don't have the message id right away
		// get this user from their email, and look for the last email sent to them that has no message id
		user, err := this.UserFromEmail (ctx, userEmail)
		if user == nil || err != nil { return err } // ignore, we couldn't find them

		// we found them, now look for their last email
		err = this.DB.QueryRow (ctx, `SELECT id, status, message_id FROM emails 
								WHERE "user" = $1 ORDER BY sent_time DESC LIMIT 1`, 
								user.Id).Scan(&email.Id, &email.Status, &email.MessageId)
		if this.ErrNoRows (err) { 
			return errors.Errorf("got a webhook about an email we've never messaged : %s : %s", email, messageId)
		}

		// we expect this to be empty
		if email.MessageId.Valid() {
			return errors.Errorf("we got a webhook from a missing message but the last user had a message id set: %s : %s : %s", userEmail, messageId, email.MessageId)
		}

		// update this email's message id for next time
		email.MessageId.Set(messageId)
		if err := this.Exec (ctx, nil, `UPDATE emails SET message_id = $2 WHERE id = $1`, email.Id, email.MessageId); err != nil {
			return err 
		}

	} else if err != nil { 
		return err // another error happened
	}

	// this have a specific priority
	if status == EmailStatus_spamreport {
		return this.emailSetStatus (ctx, email, EmailStatus(status)) // this always wins
	} else if email.Status == EmailStatus_spamreport {
		return nil // we're done
	}

	if status == EmailStatus_unsubscribe {
		return this.emailSetStatus (ctx, email, EmailStatus(status)) // this always wins
	} else if email.Status == EmailStatus_unsubscribe {
		return nil // we're done
	}

	if status == EmailStatus_groupUnsub {
		return this.emailSetStatus (ctx, email, EmailStatus(status)) // this always wins
	} else if email.Status == EmailStatus_groupUnsub {
		return nil // we're done
	}

	if status == EmailStatus_dropped {
		return this.emailSetStatus (ctx, email, EmailStatus(status)) // this always wins
	} else if email.Status == EmailStatus_dropped {
		return nil // we're done
	}

	if status == EmailStatus_click {
		return this.emailSetStatus (ctx, email, EmailStatus(status)) // this always wins
	} else if email.Status == EmailStatus_click {
		return nil // we're done
	}

	if status == EmailStatus_open {
		return this.emailSetStatus (ctx, email, EmailStatus(status)) // this always wins
	} else if email.Status == EmailStatus_open {
		return nil // we're done
	}

	if status == EmailStatus_delivered {
		return this.emailSetStatus (ctx, email, EmailStatus(status)) // this always wins
	} else if email.Status == EmailStatus_delivered {
		return nil // we're done
	}

	if status == EmailStatus_deferred {
		return this.emailSetStatus (ctx, email, EmailStatus(status)) // this always wins
	} else if email.Status == EmailStatus_deferred {
		return nil // we're done
	}

	if status == EmailStatus_processed {
		return this.emailSetStatus (ctx, email, EmailStatus(status)) // this always wins
	} else if email.Status == EmailStatus_processed {
		return nil // we're done
	}

	return errors.Errorf("uknown email status: %s", status)
}