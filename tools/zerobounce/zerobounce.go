/** ****************************************************************************************************************** **
	knows how to validate emails via zero bounce api
	
** ****************************************************************************************************************** **/

package zerobounce 

import (
	"coldbrew/tools"
	
	"github.com/pkg/errors"

	"fmt"
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

	resp, err := tools.MicroSend (ctx, http.MethodGet, 
		fmt.Sprintf("https://api.zerobounce.net/v2/validate?api_key=%s&verify_plus=false&email=%s", apiToken, email), 
		make(http.Header), make(url.Values), nil, out)
	if err != nil {
		return false, errors.Wrapf(err, "%s", string(resp))
	}

	return out.isValid(), nil
}
