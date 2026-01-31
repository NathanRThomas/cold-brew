/** ****************************************************************************************************************** **
	SQL queries related to the templates table
	
** ****************************************************************************************************************** **/

package postgres

import (
	"coldbrew/tools"
	"coldbrew/db"
	
	"github.com/google/uuid"
	"github.com/pkg/errors"
	
	"context"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type TemplateMask int64
const (
	TemplateMask_deleted 			TemplateMask = 1 << iota 
	TemplateMask_paused
	TemplateMask_warmup
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type Template struct {
	db.DBStruct
	Html, Body, Subject, Preview tools.String
	Attr struct {
		
	}
	Mask TemplateMask
}

func (this *Template) Key () string {
	return this.PrefixKey("template")
}

func (this *Template) CacheTime () tools.TimeDuration {
	return tools.TimeDuration(600) // this can cache for a while
}

func (this *Template) GenerateTextBody () (string, error) {
	return tools.GenText ("email", this.Body.String(), nil)
}

func (this *Template) GenerateHTMLBody () (string, error) {
	return tools.GenText ("email", this.Html.String(), nil)
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

func (this *Coldbrew) Template (ctx context.Context, templateId *uuid.UUID) (*Template, error) {
	template := &Template{}
	err := this.DB.QueryRow (ctx, `SELECT id, body_html, body_text, subject, preview_text, attr, mask FROM templates 
									WHERE id = $1`, templateId).Scan (&template.Id, &template.Html, 
										&template.Body, &template.Subject, &template.Preview, &template.Mask)
	
	if this.ErrNoRows (err) { return nil, nil }
	return template, errors.WithStack(err)
}

// lists all the non-paused templates
func (this *Coldbrew) TemplateList (ctx context.Context) ([]*Template, error) {
	
	rows, err := this.DB.Query (ctx, `SELECT id, body_html, body_text, subject, preview_text, attr, mask
								FROM templates WHERE mask & $1 = 0`, 
								TemplateMask_deleted | TemplateMask_paused)
	if err != nil { return nil, errors.WithStack(err) }
	defer rows.Close()

	ret := make([]*Template, 0, 3)
	for rows.Next() {
		template := &Template{}
		err := rows.Scan(&template.Id, &template.Html, &template.Body, &template.Subject, &template.Preview, &template.Attr, &template.Mask)
		if err != nil { return nil, errors.WithStack (err) }

		ret = append (ret, template)
	}

	return ret, nil
}
