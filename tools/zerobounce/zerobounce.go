/** ****************************************************************************************************************** **
	knows how to validate emails via zero bounce api
	
** ****************************************************************************************************************** **/

package zerobounce 

import (
	"coldbrew/tools"
	
	"github.com/pkg/errors"

	"net/http"
	"net/url"
	"strings"
	"context"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type zeroBounceResponse struct {
	Status string
}

func (this *zeroBounceResponse) isValid() bool {
	return strings.EqualFold (this.Status, "valid")
}


  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// generic function to help send transactional emails through sendgrid
func ValidateEmail (ctx context.Context, apiToken, email string) (bool, error) {
	out := &zeroBounceResponse{}

	params := make(url.Values)
	params.Add("api_key", apiToken)
	params.Add("verify_plus", "false")
	params.Add("email", email)

	resp, err := tools.MicroSend (ctx, http.MethodGet, 
		"https://api.zerobounce.net/v2/validate", make(http.Header), params, nil, out)
	if err != nil {
		return false, errors.Wrapf(err, "%s", string(resp))
	}

	return out.isValid(), nil
}
