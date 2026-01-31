/** ****************************************************************************************************************** **
	Shared command level stuff for all the binaries we can build
	
** ****************************************************************************************************************** **/

package cmd 

import (
	"coldbrew/db"
	"coldbrew/tools"
	"coldbrew/tools/logging"
	
	"github.com/pkg/errors"
	"github.com/google/uuid"
	json "github.com/json-iterator/go"
	
	"os"
	"os/signal"
	"syscall"
	"time"
	"math/rand"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

var (
	ContextTimeout 	tools.TimeDuration = 50 // seconds for a request to finish, this is so everything finishes < 60 seconds which is the kubernetes timeout
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CFG -------------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type CFG struct {
	Port string
	Production bool 
	ApiUrl string
	Coldbrew db.SqlCFG
}

// for parsing command line arguments
type OPTS struct {
	Help bool `short:"h" long:"help" description:"Shows help message"`
	ConfigFile string `long:"config" description:"Sets the name and location for the config file to use"`
	Production bool `long:"production" description:"Indicates this is running in a production enviroment"`
	Port string `short:"p" long:"port" default:"8080" description:"Specifies the target port to run on"`
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- APP -------------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type App struct {
	logging.Logger
	Running bool // flag for letting other threads know the parent is shutting down
}


//----- FUNCTIONS ---------------------------------------------------------------------------------------------------------//

type Callback = func() error // generic callback function that returns an error

// Base init object, so we can handle any initialize things for our App
// as well as a callback function to close down anything created here
func (this *App) Init (cfg CFG) (Callback, error) {
	// default logger for formatted error messages
	this.Logger.Init()

	rand.Seed(time.Now().UnixNano()) // for our random numbers

	uuid.EnableRandPool() // call this to help with uuid randomness

	this.Running = true // never a reason not to set this

	return EmptyCallback, nil
}

// monitors for a kill sigterm to set the running = false
// fires a custom function when it exits
func (this *App) MonitorForKill() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c // this sits until something comes into the channel, eg the notify interupts from above
	this.Running = false
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS ------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

func ParseConfig (cfg interface{}, cfgLoc string) error {
	if len(cfgLoc) == 0 {
		cfgLoc = os.Getenv("CONFIG") // default to the env variable

		if len(cfgLoc) == 0 { return nil } // they don't need a config file for this adventure
	}

	config, err := os.Open(cfgLoc)
	if err != nil { return errors.WithStack (err) }

	jsonParser := json.NewDecoder (config)
	return errors.WithStack(jsonParser.Decode (cfg))
}

// I don't like having to check for a nil callback function so i created this 
func EmptyCallback () error {
	return nil 
}
