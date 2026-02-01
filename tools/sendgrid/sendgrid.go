/** ****************************************************************************************************************** **
	knows how to send emails via sendgrid api
	
** ****************************************************************************************************************** **/

package sendgrid 

import (
	"coldbrew/tools"
	
	"github.com/pkg/errors"

	"net/http"
	"net/url"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type sendgridContent struct {
	Type string `json:"type"`
	Value string `json:"value"`
}

type sendgridUser struct {
	Email string `json:"email"`
	Name string `json:"name"`
}

type sendgridPersonalization struct {
	To []sendgridUser `json:"to"`
	Data map[string]string `json:"dynamic_template_data"`
}

type sendgridTracking struct {
	Click struct {
		Enable bool `json:"enable"`
	} `json:"click_tracking"`
	Open struct {
		Enable bool `json:"enable"`
	} `json:"open_tracking"`
	
}


  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// generic function to help send transactional emails through sendgrid
func SendEmail (apiToken, email, category, subject, textBody, htmlBody, pool, fromEmail, fromName, replyName, replyEmail string) error {
	ctx, cancel := tools.TimeDuration (60).Context ("sendEmail")
	defer cancel()

	header := make(http.Header)
    header.Add("Content-Type", "application/json")
    header.Add("Authorization", "Bearer " + apiToken)

	var req struct {
		Personalization []sendgridPersonalization `json:"personalizations"`
		From sendgridUser `json:"from"`
		Reply sendgridUser `json:"reply_to"`
		Subject string `json:"subject"`
		Tracking sendgridTracking `json:"tracking_settings"`
		Content []sendgridContent `json:"content"`
		Categories []string `json:"categories"`
		Pool string `json:"ip_pool_name"`
	}

	req.From.Email = fromEmail
	req.From.Name = fromName
	req.Reply.Email = replyEmail
	req.Reply.Name = replyName
	req.Pool = pool
	req.Subject = subject
	req.Tracking.Open.Enable = true 
	req.Tracking.Click.Enable = true 
	req.Categories = append (req.Categories, category)

	if len(textBody) > 0 {
		req.Content = append (req.Content, sendgridContent {
			Type: "text/plain",
			Value: textBody,
		})
	}
	if len(htmlBody) > 0 {
		req.Content = append (req.Content, sendgridContent {
			Type: "text/html",
			Value: htmlBody,
		})
	}

	per := sendgridPersonalization{}
	per.To = append(per.To, sendgridUser {
		Email: email,
	})

	req.Personalization = append (req.Personalization, per)
	
	resp, err := tools.MicroSend (ctx, http.MethodPost, "https://api.sendgrid.com/v3/mail/send", header, make(url.Values), req, nil)
	if err != nil {
		return errors.Wrapf(err, "%s", string(resp))
	}

	return nil 
}
