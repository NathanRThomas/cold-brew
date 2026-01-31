/** ****************************************************************************************************************** **
	Base sms class to build off.

	Started with Twilio but we can add others here as well
	
** ****************************************************************************************************************** **/

package tools 

import (
	"github.com/pkg/errors"
	
	"context"
	"strings"
	"net/url"
	"net/http"
	"fmt"
	"io"
	json "github.com/json-iterator/go"
)

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- CONST -------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCT ------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type Twilio struct {
	AccountId, Auth, VerifyServiceSms, VerifyServiceEmail string
	
}

//----- PUBLIC -----------------------------------------------------------------------------------------------------------//

// using twilio's login/verify service
func (this *Twilio) SendVerify (ctx context.Context, to string, sms bool) error {

	data := url.Values{}
	if sms {
    	data.Set("To", "+" + to)
		data.Set("Channel", "sms")
	} else {
		data.Set("To", to)
		data.Set("Channel", "email")
	}
    

	verifyService := this.VerifyServiceEmail
	if sms { verifyService = this.VerifyServiceSms }

	client := &http.Client{}
	req, err := http.NewRequestWithContext (ctx, http.MethodPost, 
					fmt.Sprintf("https://verify.twilio.com/v2/Services/%s/Verifications", verifyService), 
					strings.NewReader(data.Encode()))
	if err != nil { return errors.WithStack (err) }

	req.SetBasicAuth(this.AccountId, this.Auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil { return errors.WithStack (err) }

	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		body, _ := io.ReadAll (resp.Body)

		return errors.Errorf("twilio send verify failed : %d : %s : %s : %s", resp.StatusCode, to, string(body), verifyService)
	}

	return nil // we're good
}

// verifies the code the user thinks they got back
func (this *Twilio) VerifyCode (ctx context.Context, to, code string, sms bool) (bool, error) {
	data := url.Values{}
	if sms {
    	data.Set("To", "+" + to)
	} else {
		data.Set("To", to)
	}
    
    data.Set("Code", code)

	verifyService := this.VerifyServiceEmail
	if sms { verifyService = this.VerifyServiceSms }

	client := &http.Client{}
	req, err := http.NewRequestWithContext (ctx, http.MethodPost, 
					fmt.Sprintf("https://verify.twilio.com/v2/Services/%s/VerificationCheck", verifyService), 
					strings.NewReader(data.Encode()))
	if err != nil { return false, errors.WithStack (err) }

	req.SetBasicAuth(this.AccountId, this.Auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil { return false, errors.WithStack (err) }

	defer resp.Body.Close()

	if resp.StatusCode == 400 || resp.StatusCode == 404 {
		// means they got the code wrong
		return false, nil 

	} else if resp.StatusCode == 429 {
		return false, nil // they failed too many times to try again

	} else if resp.StatusCode < 300 {
		// we need to look at the status to see if it'as approved
		body, _ := io.ReadAll (resp.Body)
		var ret struct {
			Status string
		}

		err = json.Unmarshal(body, &ret)
		if strings.EqualFold(ret.Status, "approved") == false {
			return false, nil // this wasn't the correct code
		}

	} else { // something else bad happened
		body, _ := io.ReadAll (resp.Body)

		return false, errors.Errorf("twilio verify failed : %d : %s : %s", resp.StatusCode, string(body), verifyService)
	}

	return true, nil // we're good
}
