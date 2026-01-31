/** ****************************************************************************************************************** **
	Main API for Coldbrew
	
** ****************************************************************************************************************** **/

package main 

import (
	"coldbrew/tools"
	"coldbrew/cmd"
	"coldbrew/db/postgres"
	"coldbrew/pkg/api"
	
	"github.com/jessevdk/go-flags"
	
	"fmt"
	"os"
	"log"
	"log/slog"
	"strings"
	"time"
)

  //-------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------//

const serviceVersion = "0.1.0"
const serviceName = "Coldbrew API"

  //-------------------------------------------------------------------------------------------------------------------//
 //----- CONFIG ------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------//


// final local options object for this executable
var opts struct {
	cmd.OPTS
}

var cfg struct {
	cmd.CFG
	AdminToken string
}


  //-------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------//
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
	
	api *api.API 
	
	db 			*postgres.Coldbrew
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
	coldbrewDB, err := cfg.Coldbrew.Connect()
	if err != nil {
		return func() error {
			return cmdFn()
		}, err
	}

	this.db = postgres.NewColdbrew(coldbrewDB)
	
	if err == nil {
		this.api = api.NewAPI(this.db, tools.String(cfg.ApiUrl), cfg.Production)
	}
	
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

	// this means we're exiting
	slog.Info("API exiting")
	time.Sleep (time.Second * 5) // for being removed from the load balancer, we need time to become "unhealthy"
	
	// ok, we're done with the server
	gg.Shutdown()

	app.StackTrace(nil, finalDefer())
	
	os.Exit(0) //final exit
}
