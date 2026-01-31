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
	EmailStatus_bounced		= EmailStatus("bounced")
	EmailStatus_blocked		= EmailStatus("blocked")
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

// lists all the non-paused emails
func (this *Coldbrew) EmailInsert (ctx context.Context, email *Email) error {
	email.SetPK()

	return this.Exec (ctx, nil, `INSERT INTO emails (id, target_time, mailman, template, "user", status) 
									VALUES ($1, $2, $3, $4, $5, $6)`, email.Id, email.Target, 
									email.Mailman, email.Template, email.User, email.Status)
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
