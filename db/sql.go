/** ****************************************************************************************************************** **
	Base sql db class.  This should be extended in higher level applications.

** ****************************************************************************************************************** **/

package db 

import (
	"coldbrew/tools"

	"github.com/jackc/pgx/v5/pgxpool" 
	"github.com/jackc/pgx/v5"
    "github.com/pkg/errors" 

	"context"
	"strings"
	"time"
	"fmt"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

//----- SqlCFG -------------------------------------------------------------------------------------------------------------//

type SqlCFG struct {
	IP, Database, User, Password tools.String  
	Port int
}

func (this *SqlCFG) valid () error {
	if this.IP.Valid() == false { return errors.Errorf ("No ip listed for db") }
	if this.Database.Valid() == false { return errors.Errorf ("No database specified") }
	if this.Password.Valid() == false { return errors.Errorf ("No password found for the database: %s", this.Database) }
	if this.Port == 0 { this.Port = 5432 } // set some defaults
	if this.User.Valid() == false { this.User.Set("root") }

	return nil
}

func (this *SqlCFG) Connect () (*pgxpool.Pool, error) {
	err := this.valid()
	if err != nil { return nil, err }
	
	sslmode := "sslmode=require"

	var cockDB *pgxpool.Pool 
	var errOut error

	ctx, cancel := context.WithTimeout (context.Background(), time.Second * 10)
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?%s", this.User, this.Password, this.IP, this.Port, this.Database, sslmode)

	config, err := pgxpool.ParseConfig (connStr)
	if err != nil {
		return nil, errors.Wrapf (err, "ip : %s", this.IP)
	}

	go func() {
		cockDB, err = pgxpool.NewWithConfig (ctx, config)

		if err != nil && errOut == nil { errOut = err }
		if err == nil { cancel() } // we're good
	}()

	select {
	case <-ctx.Done(): // good
		return cockDB, nil 

	case <-time.After(time.Second * 9): // bad
		cancel()
		if errOut == nil { errOut = errors.Errorf ("connection timed out : %s", this.IP) }

	}
	
	return nil, errOut
}

type SQL struct {
	DB *pgxpool.Pool
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

func (this *SQL) Exec (ctx context.Context, tx pgx.Tx, query string, params ...interface{}) (err error) {
	commit := false 

	if tx == nil {
		commit = true
		tx, err = this.DB.Begin (ctx) // create a local transaction instead
		
		if err != nil {
			return errors.WithStack(err)
		}

		defer tx.Rollback(ctx) // rollback only the local one
	}
	

	_, err = tx.Exec (ctx, query, params...)
	if err == nil {
		if commit { // only commit if we created a tx locally
			err = errors.WithStack(tx.Commit(ctx))
		}
	} else {
		err = errors.WithStack(err)
	}
	return
}

func (this *SQL) ErrNoRows (err error) bool {
	if err == nil { return false }
	return strings.Index (err.Error(), "no rows in result set") >= 0
}

func (this *SQL) ErrUniqueConstraint (err error) bool {
	if err == nil { return false }
	return strings.Index (err.Error(), "duplicate key value violates unique constraint") >= 0
}

// wrapper function around the above to return the error of interest
func (this *SQL) ErrNotFound (err error) error {
	if this.ErrNoRows(err) {
		return errors.WithStack (ErrKeyNotFound)
	}
	return errors.WithStack(err) // else just return what we got
}

func (this *SQL) Ping (ctx context.Context) error {
	conn, err := this.DB.Acquire (ctx)
	if err == nil { conn.Release() }
	return errors.WithStack (err)
}

func (this *SQL) HealthCheck (ctx context.Context) error {
	sum := 0
	err := this.DB.QueryRow (ctx, `SELECT 1+1`).Scan(&sum)
	if err != nil { return errors.WithStack(err) }
	if sum != 2 { return errors.Errorf("Math doesn't add up") }
	return nil // we're good
}

func (this *SQL) Begin (ctx context.Context) (pgx.Tx, error) {
	tx, err := this.DB.Begin(ctx)
	return tx, errors.WithStack(err)
}
