/** ****************************************************************************************************************** **
	
	
** ****************************************************************************************************************** **/

package postgres

import (
	"coldbrew/db"

    "github.com/jackc/pgx/v5/pgxpool" 

    "time"
    "fmt"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type Coldbrew struct {
	*db.SQL
}

type mailmanPerformance struct {
	Cnt int
	Status EmailStatus
	Sent time.Time
}

type mailmanPerformances []*mailmanPerformance

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

func (this mailmanPerformances) duration () time.Duration {
    // see how many dates we have
    // these should already be in order of date
    if len(this) < 2 { return time.Hour } // we don't have any data, so send them once an hour

    days := int(this[0].Sent.Sub(this[len(this)-1].Sent).Hours()) / 24 // covner this difference into days
    fmt.Println("mailman performance: days", days)

    if days < 5 { return time.Minute * 40 } // less than 5 days, a little faster, ~50/day

    if days < 11 {
        // we have some data here, let's just increase a little, we should be able to do 100 or so per day even if they're "bad"
        // 100 / 24 ~ 4
        return time.Minute * 30 // so 1 every 15 mintues ~ 100/day
    }

    // let's have some fun, let's look at overall sends, compared to "bad" things
    sends := 0
    bads := 0
    good := 0
    for _, mp := range this {
        switch mp.Status {
        case EmailStatus_processed, "":
            sends += mp.Cnt 

        case EmailStatus_delivered: // this is good
            good += mp.Cnt

        default:
            // the rest are bad
            bads += mp.Cnt
        }
    }

    fmt.Println("mailman performance: ", sends, bads, good)
    if sends < 1 {
        fmt.Println("WE HAD NO SENDS")
        return time.Hour
    }

    if float64(bads) / float64(sends) > 0.15 || good < 60 {
        fmt.Println("mailman performance: bad sends ratio", float64(bads) / float64(sends))
        return time.Minute * 15 // (60 / 15) * 24 = 96 slow it down
    }

    if float64(bads) / float64(sends) > 0.1 || good < 100 {
        fmt.Println("mailman performance: bad sends ratio", float64(bads) / float64(sends))
        return time.Minute * 12 // (60 / 12) * 24 = 120
    }

    if float64(bads) / float64(sends) > 0.5 || good < 200 {
        fmt.Println("mailman performance: bad sends ratio", float64(bads) / float64(sends))
        return time.Minute * 6 // (60 / 6) * 24 = 240
    }

    if float64(bads) / float64(sends) > 0.02 || good < 400 {
        fmt.Println("mailman performance: bad sends ratio", float64(bads) / float64(sends))
        return time.Minute * 3 // (60 / 3) * 24 = 480
    }

    if float64(bads) / float64(sends) > 0.01 {
        fmt.Println("mailman performance: bad sends ratio", float64(bads) / float64(sends))
        return time.Minute * 2 // (60 / 2) * 24 = 720
    }

    // full throttle
    fmt.Println("mailman performance: full throttle")
    return time.Minute // = 1440
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

func NewColdbrew (d *pgxpool.Pool) *Coldbrew {
	return &Coldbrew {
		SQL: &db.SQL {
			DB: d,
		},
	}
}
