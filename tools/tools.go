/** ****************************************************************************************************************** **
	Re-used functions, often related to our custom String object
	
** ****************************************************************************************************************** **/

package tools 

import (
	"coldbrew/tools/logging"

	"github.com/pkg/errors"

	"fmt"
	"context"
	"math"
	json "github.com/json-iterator/go"
	"time"
	"os"
	"bytes"
	"text/template"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// full copy of a struct including the pointers
func DeepCopy (in, out interface{}) error {
	jstr, err := json.Marshal (in)
	if err != nil { return errors.WithStack(err) }

	err = json.Unmarshal(jstr, out)
	return errors.WithStack(err)
}


//----- TYPES -------------------------------------------------------------------------------------------------------//

type TimeDuration int // a way for us to set timeouts and get the duration easier

func (this TimeDuration) Duration () time.Duration {
	return time.Duration(this) * time.Second 
}

func (this TimeDuration) Sleep () { // sleeper function
	time.Sleep(this.Duration())
}

// redis wants this as a string as an int
func (this TimeDuration) Redis () string {
	return fmt.Sprintf("%.0f", this.Duration().Seconds())
}

func (this TimeDuration) Context (threadName string) (context.Context, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), this.Duration())

	if len(threadName) > 0 {
		ctx = context.WithValue(ctx, logging.ThreadName, threadName) // add this to our context
	}

	ctx = context.WithValue(ctx, logging.ThreadId, time.Now().Format("2006-01-02 15:04:05")) // add this to our context

	return ctx, cancel
}

func (this TimeDuration) Ticker () *time.Ticker {
	return time.NewTicker(this.Duration())
}

// converts this into a more friendly readout
func (this TimeDuration) CountDown () string {
	if this.Duration().Hours() >= 48 { // more than 2 day
		return fmt.Sprintf ("%.0f days", this.Duration().Hours() / 24)
	} 
	if this.Duration().Hours() >= 24 { // more than 1 day
		return fmt.Sprintf ("%.0f day", math.Floor(this.Duration().Hours() / 24))
	} 
	if this.Duration().Hours() > 1 {
		return fmt.Sprintf ("%.0f hours", this.Duration().Hours())
	}
	if this.Duration().Minutes() > 1 {
		return fmt.Sprintf ("%.0f minutes", this.Duration().Minutes())
	}
	if this.Duration().Seconds() > 1 {
		return fmt.Sprintf ("%.0f seconds", this.Duration().Seconds())
	}
	if this.Duration().Seconds() > 0 {
		return "1 second"
	}

	return "now"
}

func GenText (dir, filename string, data interface{}) (string, error) {
	loc := os.Getenv ("TEMPLATES")
	if len(loc) == 0 { return "", errors.Errorf ("environment variable TEMPLATES not set") }

	if len(dir) > 0 {
		loc += "/" + dir
	}
	loc += "/" + filename // final location

	txt, err := template.New(filename).ParseFiles(loc) // get the text version
	if err != nil { return "", errors.Wrapf (err, "template not found : %s", loc) }

	txtBuf := new(bytes.Buffer)

	err = txt.Execute (txtBuf, data)
	if err != nil { 
		str, _ := json.Marshal (data)
		return "", errors.Wrapf (err, "%s :: %s\n", filename, str) 
	}

	return txtBuf.String(), nil 
}

