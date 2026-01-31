/** ****************************************************************************************************************** **
	Tools used for testing, so we can include them in any _test.go file

** ****************************************************************************************************************** **/

package tools

import (
	"coldbrew/tools/logging"

	"github.com/pkg/errors"

	"testing"
	"os"
	json "github.com/json-iterator/go"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// used during testing so we can see where the unexpected error is coming from 
func TestingStackTrace (t *testing.T, err error) {
	if err == nil { return }
	for _, ln := range logging.StackTraceToArray (err) {
		t.Logf("%s\n", ln)
	}

	t.Fatal(err)
}

// reads the default testing config file
func TestingLoadConfig (cfg interface{}) error {
	config, err := os.Open("./testing/config.json") // hard-coded so we do need to run tests from the parent dir not on windows
	if err != nil { 
		// assume we're in the wrong directory
		config, err = os.Open("./../testing/config.json")
		if err != nil {
			config, err = os.Open("./../../testing/config.json")
			if err != nil {
				config, err = os.Open("./../../../testing/config.json")
				if err != nil {
					config, err = os.Open("./../../../../testing/config.json")
				}
			}
		}
	}
	if err != nil { return errors.WithStack(err) } // we couldn't find the config file

	jsonParser := json.NewDecoder (config)
	return errors.WithStack (jsonParser.Decode (cfg))
}

func TestingAssertRange (t *testing.T, min, max, val float64) {
	if val < min {
		t.Fatalf("%.2f is outside the min range %.2f", val, min)
	}
	if val > max {
		t.Fatalf("%.2f is outside the max range %.2f", val, max)
	}
}

func TestingAssertFloat (t *testing.T, target, val float64) {
	TestingAssertRange (t, target * .99, target * 1.01, val)
}
