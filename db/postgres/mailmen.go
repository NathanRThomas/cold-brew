/** ****************************************************************************************************************** **
	SQL queries related to the mailmen table
	
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

type MailmanMask int64
const (
	MailmanMask_deleted 			MailmanMask = 1 << iota 
	MailmanMask_textWarm
	MailmanMask_htmlWarm
	MailmanMask_paused
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type Mailman struct {
	db.DBStruct
	Attr struct {
		IpPool, FromEmail, FromName, ReplyEmail, ReplyName, Category, APIToken tools.String
	}
	Mask MailmanMask
}

func (this *Mailman) Key () string {
	return this.PrefixKey("mailman")
}

func (this *Mailman) CacheTime () tools.TimeDuration {
	return tools.TimeDuration(600) // this can cache for a while
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

func (this *Coldbrew) Mailman (ctx context.Context, mailmanId *uuid.UUID) (*Mailman, error) {
	mailman := &Mailman{}
	err := this.DB.QueryRow (ctx, `SELECT id, attr, mask FROM mailmen 
									WHERE id = $1`, mailmanId).Scan (&mailman.Id, &mailman.Attr, &mailman.Mask)
	
	if this.ErrNoRows (err) { return nil, nil }
	return mailman, errors.WithStack(err)
}

// lists all the non-paused mailman
func (this *Coldbrew) MailmanList (ctx context.Context) ([]*Mailman, error) {
	
	rows, err := this.DB.Query (ctx, `SELECT id, attr, mask FROM mailmen WHERE mask & $1 = 0`, 
								MailmanMask_deleted | MailmanMask_paused)
	if err != nil { return nil, errors.WithStack(err) }
	defer rows.Close()

	ret := make([]*Mailman, 0, 3)
	for rows.Next() {
		mm := &Mailman{}
		err := rows.Scan(&mm.Id, &mm.Attr, &mm.Mask)
		if err != nil { return nil, errors.WithStack (err) }

		ret = append (ret, mm)
	}

	return ret, nil
}

// updates the mask for a mailman
func (this *Coldbrew) MailmanSetMask (ctx context.Context, mailman *Mailman, mask MailmanMask) error {
	if mailman.Mask & mask == mask { return nil } // already good
	mailman.Mask |= mask // update it in real-time
	
	return this.Exec (ctx, nil, `UPDATE mailmen SET mask = mask | $1 WHERE id = $2`, mask, mailman.Id)
}

// figures out the sending frequency based on past performance
func (this *Coldbrew) MailmanPerformance (ctx context.Context, mailmanId *uuid.UUID) (time.Duration, error) {
	rows, err := this.DB.Query (ctx, `SELECT COUNT(*), status, sent_time::date as sent 
									FROM emails WHERE mailman = $1 AND sent_time IS NOT NULL 
									GROUP BY sent, status ORDER BY sent DESC LIMIT 150`, mailmanId)
	if err != nil { return 0, errors.WithStack(err) }
	defer rows.Close()

	performance := make(mailmanPerformances, 0, 150) // we just need to go back far enough to get a trend line and history

	for rows.Next() {
		p := &mailmanPerformance{}
		err := rows.Scan(&p.Cnt, &p.Status, &p.Sent)
		if err != nil { return 0, errors.WithStack(err) }

		performance = append (performance, p)
	}

	return performance.duration(), nil 
}