/** ****************************************************************************************************************** **
	SQL queries related to the users table
	
** ****************************************************************************************************************** **/

package postgres

import (
	"coldbrew/tools"
	"coldbrew/db"
	
	"github.com/google/uuid"
	"github.com/pkg/errors"
	
	"time"
	"fmt"
	"context"
	"crypto/sha256"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type UserMask int64
const (
	UserMask_deleted 			UserMask = 1 << iota 
	UserMask_warmup 			// indicates this user is used to warm up new mailmen
	_
	UserMask_delivered

	UserMask_open
	UserMask_click
	UserMask_dropped
	UserMask_failed

	UserMask_unsubscribe
	UserMask_deferred
	UserMask_bounce
	UserMask_spam
)

const UserMask_doNotEmail = UserMask_deleted | UserMask_unsubscribe | UserMask_failed | UserMask_dropped | UserMask_bounce | UserMask_spam

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type User struct {
	db.DBStruct
	Email, Token tools.String
	Attr struct {
		
	}
	Mask UserMask
}

func (this *User) Key () string {
	return this.PrefixKey("user")
}

func (this *User) CacheTime () tools.TimeDuration {
	return tools.TimeDuration(60)
}

func (this *User) init () {
	this.SetPK()
	this.GenerateToken() // give them a token
}

// generates a new sha256 token for this user
func (this *User) GenerateToken() {
	h := sha256.New()
    h.Write([]byte(fmt.Sprintf("%s%d:%s:%s", db.Salt, time.Now().UnixNano(), this.Id.String(), this.Email)))
    this.Token.Sprintf("%x", h.Sum(nil))
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

func (this *Coldbrew) User (ctx context.Context, userId *uuid.UUID) (*User, error) {
	user := &User{}
	err := this.DB.QueryRow (ctx, `SELECT id, email, token, attr, mask FROM users 
									WHERE id = $1`, userId).Scan (&user.Id, &user.Email, &user.Token, &user.Attr, &user.Mask)
	
	if this.ErrNoRows (err) { return nil, nil }
	return user, errors.WithStack(err)
}

func (this *Coldbrew) UserInsert (ctx context.Context, warmup bool, email tools.String) error {
	user := &User {
		Email: email,
	}
	user.init()
	if warmup { user.Mask |= UserMask_warmup }

	err := this.Exec (ctx, nil, `INSERT INTO users (id, email, token, mask) VALUES ($1, $2, $3, $4)`,
						user.Id, user.Email, user.Token, user.Mask)
	if this.ErrUniqueConstraint (err) { return nil } // don't record this error
	return err
}

// finds the user based on their bearer token
func (this *Coldbrew) UserFromBearer (ctx context.Context, bearer tools.String) (*User, error) {
	var exists *uuid.UUID
	err := this.DB.QueryRow (ctx, `SELECT id FROM users WHERE token = $1`, bearer).Scan (&exists)
	
	if this.ErrNoRows (err) { return nil, nil }
	if err != nil { return nil, errors.WithStack(err) }

	return this.User (ctx, exists)
}

func (this *Coldbrew) UserFromEmail (ctx context.Context, email tools.String) (*User, error) {
	var exists *uuid.UUID
	err := this.DB.QueryRow (ctx, `SELECT id FROM users WHERE email = $1`, email).Scan (&exists)
	
	if this.ErrNoRows (err) { return nil, nil }
	if err != nil { return nil, errors.WithStack(err) }

	return this.User (ctx, exists)
}

// updates the mask for a user
func (this *Coldbrew) UserSetMask (ctx context.Context, user *User, mask UserMask) error {
	if user.Mask & mask == mask { return nil } // already good
	user.Mask |= mask // update it in real-time
	
	if err := this.Exec (ctx, nil, `UPDATE users SET mask = mask | $1 WHERE id = $2`, mask, user.Id); err != nil { return err }

	// that worked, see if they should be disabled
	if user.Mask & UserMask_doNotEmail > 0 {
		return this.UserSetDisabled (ctx, user)
	}
	return nil // we're good
}

// removes a mask from the user
func (this *Coldbrew) UserRemoveMask (ctx context.Context, user *User, mask UserMask) error {
	if user.Mask & mask == 0 { return nil } // already good
	user.Mask = user.Mask & ^ mask // update it in real-time
	
	return this.Exec (ctx, nil, fmt.Sprintf(`UPDATE users SET mask = mask & (~%d) WHERE id = $1`, mask), user.Id)
}

func (this *Coldbrew) UsersFromMask (ctx context.Context, mask UserMask) ([]*User, error) {
	rows, err := this.DB.Query (ctx, `SELECT id, email, attr FROM users WHERE mask & $1 = $2`, 
								mask | UserMask_doNotEmail, mask)
	if err != nil { return nil, errors.WithStack(err) }
	defer rows.Close()

	ret := make([]*User, 0, 128)

	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.Id, &user.Email, &user.Attr)
		if err != nil { return nil, errors.WithStack(err) }

		ret = append (ret, user)
	}

	return ret, nil
}

// finds a user that needs to be validated
func (this *Coldbrew) UserNotValidated (ctx context.Context) (*User, error) {
	ret := &User{}
	err := this.DB.QueryRow (ctx, `SELECT id, email, attr FROM users WHERE validated IS NULL AND disabled IS NULL limit 1`).Scan(
		&ret.Id, &ret.Email, &ret.Attr)
	if this.ErrNoRows (err) { return nil, nil }
	return ret, errors.WithStack(err)
}

func (this *Coldbrew) UserSetValid (ctx context.Context, user *User) error {
	return this.Exec (ctx, nil, `UPDATE users SET validated = NOW() WHERE id = $1`, user.Id)
}

func (this *Coldbrew) UserSetDisabled (ctx context.Context, user *User) error {
	return this.Exec (ctx, nil, `UPDATE users SET disabled = NOW() WHERE id = $1`, user.Id)
}

// returns just a few users or so that haven't been emailed before
func (this *Coldbrew) UsersMissing (ctx context.Context) ([]*User, error) {
	rows, err := this.DB.Query (ctx, `SELECT id, email, attr 
								FROM users 
								WHERE disabled IS NULL AND validated IS NOT NULL AND id NOT IN 
								(SELECT "user" FROM emails)
								LIMIT 20`) // only include validated emails
	if err != nil { return nil, errors.WithStack(err) }
	defer rows.Close()

	ret := make([]*User, 0, 10)

	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.Id, &user.Email, &user.Attr)
		if err != nil { return nil, errors.WithStack(err) }

		ret = append (ret, user)
	}

	return ret, nil
}

// from our webhooks this updates our email status
// this might not match what's in this database, if not, then ignore it
func (this *Coldbrew) UserUpdateStatus (ctx context.Context, email tools.String, status string) error {
	user, err := this.UserFromEmail (ctx, email)
	if user == nil || err != nil { return err }
	
	// we have a user, let's update them
	switch EmailStatus(status) {
	case EmailStatus_processed:
		// don't worry about this one
		return nil
	case EmailStatus_delivered:
		return this.UserSetMask (ctx, user, UserMask_delivered)

	case EmailStatus_deferred:
		// i'm not sure what to do with this info actually
		return this.UserSetMask (ctx, user, UserMask_deferred)

	case EmailStatus_open:
		return this.UserSetMask (ctx, user, UserMask_open)

	case EmailStatus_click:
		return this.UserSetMask (ctx, user, UserMask_click)

	case EmailStatus_dropped:
		return this.UserSetMask (ctx, user, UserMask_dropped)

	case EmailStatus_bounce:
		return this.UserSetMask (ctx, user, UserMask_bounce)

	case EmailStatus_spamreport:
		return this.UserSetMask (ctx, user, UserMask_spam)

	case EmailStatus_unsubscribe, EmailStatus_groupUnsub:
		return this.UserSetMask (ctx, user, UserMask_unsubscribe)
	}

	return errors.Errorf("uknown email status: %s", status)
}
