/** ****************************************************************************************************************** **
	This is the single thread, one pod at a time, background service for the Coldbrew app

	
** ****************************************************************************************************************** **/

package main 

import (
	"coldbrew/db/postgres"
	"coldbrew/cmd"
	"coldbrew/tools"
	
	"github.com/jessevdk/go-flags"
	
	"fmt"
	"os"
	"sync"
	"strings"
	"log"
	"log/slog"
)

  //-------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------//

const serviceVersion = "0.1.0"
const serviceName = "Coldbrew QB"

  //-------------------------------------------------------------------------------------------------------------------//
 //----- CONFIG ------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------//

// final local options object for this executable
var opts struct {
	cmd.OPTS
}

var cfg struct {
	cmd.CFG

	ZeroBounce tools.String
}


  //-------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS ---------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------//

func showHelp() {
	fmt.Printf("***************************\n%s : Version %s\n\n", serviceName, serviceVersion)

	fmt.Printf("\n**************************\n")
}

// handles parsing command arguments as well as setting up our opts object
func parseCommandLineArgs() []string {
	// parse things
	args, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	if opts.Help {
		showHelp()
		os.Exit(0)
	}

	if opts.Production == true {
		cfg.Production = true 
	}
	
	// check any args
	for _, arg := range args {
		switch strings.ToLower(arg) {
		case "help":
			showHelp()
			os.Exit(0)

		case "version":
			fmt.Printf("%s\n", serviceVersion)
			os.Exit(0)
		}
	}

	return args // return any arguments we don't know what to do with... yet
}


  //-------------------------------------------------------------------------------------------------------------------//
 //----- APP ---------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------//

type app struct {
	cmd.App

	db 		*postgres.Coldbrew
}

  //-------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS ---------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------//

// local init for this app instance
func (this *app) init() (cmd.Callback, error) {
	
	// start by using an init from the base object
	cmdFn, err := this.App.Init(cfg.CFG)
	if err != nil {
		return cmdFn, err
	} // bail

	// now try to connect to our sql database
	cbDB, err := cfg.Coldbrew.Connect()
	if err != nil {
		return func() error {
			return cmdFn()
		}, err
	}

	this.db = postgres.NewColdbrew(cbDB)

	return func() error {
		this.db.DB.Close() // close our database as well

		return cmdFn()
	}, err // return any error from above
}

  //-------------------------------------------------------------------------------------------------------------------//
 //----- MAIN --------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------//

func main() {
	
	// first step, parse the command line params
	args := parseCommandLineArgs()

	// parse the config
	err := cmd.ParseConfig(&cfg, opts.ConfigFile)
	if err != nil {
		log.Fatal(err) // bailing hard
	}

	// check for a port being set
	if len(opts.Port) > 0 {
		cfg.Port = opts.Port // this wins
	}

	// early check for flags
	for _, arg := range args {
		switch strings.ToLower(arg) {
		case "production":
			cfg.Production = true // we're running in production
		}
	}

	// main app for everything
	app := &app{}

	finalDefer, err := app.init() // init our application
	if err != nil {
		app.StackTrace(nil, err)
		finalDefer()
		log.Fatal(err) // this is also super bad
	}

	// we're good to keep going

	// start some background tasks
	
	wg := new(sync.WaitGroup)

	// launch our task to check for new instant game combinations
	var emailValidateFrequency tools.TimeDuration = 3 // pretty quick validate the email address with zero bounce
	var emailSendFrequency tools.TimeDuration = 5 // pretty quick check for emails that need to be sent
	var queEmailFrequency tools.TimeDuration = 60 // once a minute, look for emails that need to get queued
	
	wg.Add(1)
	go app.FlowLaunchBlocking (wg, app.flowEmailValidate, emailValidateFrequency, cmd.ContextTimeout, "flowEmailValidate") // checks for emails that need to be validated

	wg.Add(1)
	go app.FlowLaunchBlocking (wg, app.flowEmailSend, emailSendFrequency, cmd.ContextTimeout, "flowEmailSend") // checks for emails that need to be sent

	wg.Add(1)
	go app.FlowLaunchBlocking (wg, app.flowQueEmails, queEmailFrequency, cmd.ContextTimeout, "flowQueEmails") // populates emails to be sent over the next hour

	// create our server
	gg := app.routes()

	// start the server in a background thread
	go func() {
		slog.Info(serviceName + " v" + serviceVersion + " started on port " + cfg.Port)
		if err := gg.Listen(":" + cfg.Port); err != nil {
			slog.Warn(err.Error()) // we want to know if this failed for another reason
		}
	}()

	// now wait for the cancel command
	app.MonitorForKill()
	
	slog.Info("QB exiting")
	
	wg.Wait()

	// ok, we're done with the server
	gg.Shutdown()

	app.StackTrace(nil, finalDefer())
	
	os.Exit(0) //final exit
}
