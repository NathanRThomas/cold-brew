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
	Status, Sub_status, Did_you_mean string
}

func (this *zeroBounceResponse) isValid() bool {
	return strings.EqualFold (this.Status, "valid")
}

func (this *zeroBounceResponse) isTypo() string {
	if strings.EqualFold (this.Sub_status, "possible_typo") && len(this.Did_you_mean) > 0 {
		return this.Did_you_mean
	}
	return "" // no idea
}


  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// generic function to help send transactional emails through sendgrid
func ValidateEmail (ctx context.Context, apiToken, email string) (bool, string, error) {
	out := &zeroBounceResponse{}

	params := make(url.Values)
	params.Add("api_key", apiToken)
	params.Add("verify_plus", "false")
	params.Add("email", email)

	resp, err := tools.MicroSend (ctx, http.MethodGet, 
		"https://api.zerobounce.net/v2/validate", make(http.Header), params, nil, out)
	if err != nil {
		return false, "", errors.Wrapf(err, "%s", string(resp))
	}

	// fmt.Println("response: ", string(resp))

	if out.isValid() {
		return true, "", nil // this one was good
	}
	return false, out.isTypo(), nil // see if we had a typo
}
