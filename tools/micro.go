/** ****************************************************************************************************************** **
	This is just a couple of functions that can be used to send and receive data to urls
	
** ****************************************************************************************************************** **/

package tools 

import (
	"github.com/pkg/errors"

	"context"
	json "github.com/json-iterator/go"
	"net/http"
	"net/url"
	"bytes"
	"io/ioutil"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//


// high level send, handles in and out body parsing
func MicroSend (ctx context.Context, requestType, link string, header http.Header, queryParams url.Values, in, out interface{}) ([]byte, error) {
	var jstr []byte 
	var err error 

	if in != nil {
		jstr, err = json.Marshal (in)
		if err != nil { return nil, errors.WithStack (err) }
	}

	req, err := http.NewRequestWithContext (ctx, requestType, link, bytes.NewBuffer(jstr))
	if err != nil { return nil, errors.Wrap (err, link) }

	req.URL.RawQuery = queryParams.Encode() // set our query params
	req.Header = header // set our header information

	// we're ready to actually do the request now
	resp, err := http.DefaultClient.Do (req)
	if err != nil { return nil, errors.WithStack (err) }
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll (resp.Body) // i don't think we need to check the error on this one

	if resp.StatusCode >= http.StatusBadRequest { // 400 or worse
		return respBody, errors.Errorf("ERROR %d returned from %s : %s : %+v : %s :: %s", resp.StatusCode, link, string(jstr), header, queryParams.Encode(), string(respBody))
	}

	if out != nil && resp.StatusCode >= http.StatusOK && resp.StatusCode <= http.StatusCreated { // 200 or 201
		err = errors.Wrapf (json.Unmarshal (respBody, out), " :: %s", string(respBody))
	}

	return respBody, err 
}
