/** ****************************************************************************************************************** **
	One of our core objects from tools, our String object

	A ton of built in functionality for manipulating and setting a string
	
** ****************************************************************************************************************** **/

package tools 

import (
	"github.com/google/uuid"

	"fmt"
	"strings"
	"regexp"
	"strconv"
	"net/url"
	"crypto/md5"
	"encoding/hex"

)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- TYPES -----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type String string 

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

func (this *String) clean () {
	if this == nil { return }
	this.Set (strings.TrimSpace(string(*this)))
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

func (this *String) Set (str string) {
	*this = String(str)
}

func (this *String) String () string {
	if this == nil { return "" }
	this.clean()
	return string(*this)
}

func (this *String) Sprintf (str string, params ...interface{}) {
	this.Set (fmt.Sprintf(str, params...))
}

func (this *String) Len () int {
	if this == nil { return 0 }
	return len(this.String())
}

func (this *String) Valid () bool {
	if this == nil { return false }
	return this.Len() > 0
}

func (this *String) Int () int {
	i, _ := strconv.ParseInt(this.String(), 10, 64)
	return int(i)
}

func (this *String) Email () bool {
	email := this.Remove(" ") // remove all the spaces first
	m, _ := regexp.MatchString (`^.+@.+\..+$`, email) // very generous email check
	if m {
		this.Set(email)
		return true 
	}
	return false
}

func (this *String) Equal (str string) bool {
	return strings.EqualFold (this.String(), strings.TrimSpace(str))
}

func (this *String) Remove (reg string) string {
	r, _ := regexp.Compile (reg)
	return r.ReplaceAllString(this.String(), "")
}

// checks and formats that it's a US based phone number
func (this *String) Phone () bool {
	m := regexp.MustCompile (`^[\+1 \-\(\)]*([2-9]\d{2})[\(\)\. -]{0,2}(\d{3})[\. -]?(\d{4})$`)
	r := m.FindStringSubmatch (this.String())
	if len(r) == 4 {
		this.Sprintf ("1%s%s%s", r[1], r[2], r[3])
		return true 
	}
	return false // not a good number
}

// tries to format the phone number in a human friendly form
func (this *String) PhoneFormat () {
	if this.Phone() == false { return } // not a phone number
	r := regexp.MustCompile(`^1?(\d{3})(\d{3})(\d{4})$`)
	m := r.FindStringSubmatch(this.String())
	if len(m) != 4 { return }
	this.Sprintf ("(%s) %s-%s", m[1], m[2], m[3])
}

func (this *String) IsUUID () bool {
	m := regexp.MustCompile (`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return m.Match([]byte(this.String()))
}

func (this *String) UUID () *uuid.UUID {
	u, err := uuid.Parse(this.String())
	if err != nil { return nil }
	return &u 
}

func (this *String) Url () bool {
	// make sure this isn't an email first
	if this.Email() { return false }
	
	resp, err := url.Parse (this.String())
	if err != nil { return false } // clearly not a url

	// check the scheme
	if len(resp.Scheme) == 0 {
		// you have to re-parse it with the https otherwise the host doesn't get set
		resp, err = url.Parse ("https://" + this.String()) // default to this, no one uses http anymore
		if err != nil { return false } // clearly not a url
	}

	if strings.Index (resp.Host, ".") < 1 { return false } // we don't want to allow network locaions

	this.Set (resp.String()) // we're good
	return true 
}

// converts itself to lower case
func (this *String) ToLower () {
	this.Set (strings.ToLower(this.String()))
}

func (this *String) SafeString () string {
	var n String 
	n.Set (this.Remove ("[^a-zA-Z0-9\\s]"))
	return strings.ToLower (n.String())
}

func (this *String) SuperSafeString () string {
	var n String 
	n.Set (this.Remove ("[^a-zA-Z0-9]"))
	return strings.ToLower (n.String())
}

func (this *String) MD5() string {
	algorithm := md5.New()
	algorithm.Write([]byte(this.String()))
	return hex.EncodeToString(algorithm.Sum(nil))
}

//----- ARRAY ----------------------------------------------------------------------------------------------------------//

type StringList []String 

func (this StringList) Len() int {
	return len(this)
}

// adds another appoitnment to our list
func (this *StringList) Push (str String) {
	*this = append (*this, str)
}

func (this *StringList) PushStr (in string) {
	this.Push(String(in))
}

func (this *StringList) Clear () {
	*this = StringList{}
}

// pops at a specific index
func (this *StringList) Pop (idx int) {
	l := this.Len()
	if l == 0 { return } // just a saftey check
	if idx >= l { idx = l -1 } // just pop the last one
	if idx < 0 { idx = l - 1 } // just pop the last one
	
	*this = append((*this)[:idx], (*this)[idx+1:]...)
}

// converts an array of our strings to an array of generic ones
func (this StringList) ToStrings () []string {
	var ret []string 
	for _, str := range this {
		ret = append (ret, str.String())
	}

	return ret 
}

func (this StringList) Join (sep string) string {
	return strings.Join(this.ToStrings(), sep)
}

// checks to see if the string exists in our tools.String array
func (this StringList) Exists (needle string) bool {
	for _, a := range this {
		if a.Equal (needle) {
			return true // already exists
		}
	}
	return false // this is new
}

// removes from our list if it matches
func (this *StringList) Remove (target string) {
	if this.Len() == 0 { return }

	for idx, str := range *this {
		
		if str.Equal(target) { // found it
			this.Pop(idx)
			this.Remove (target) // recurse in case there's multiple items
			// don't keep looping as we broke the range in the for loop above
			return
		}
	}

	return // didn't exist, we're good
}

// checks to see if this needle already exists in the haystack and if not adds it
func (this *StringList) PushUnique (needle string) bool {
	if this.Exists(needle) { return false } // exists already

	this.Push (String(needle)) // add it
	return true 
}

//----- FUNCTIONS ----------------------------------------------------------------------------------------------------------//

// creates a new string list by parsing the incoming string
func NewStringList (in, sep string) (ret StringList) {
	if len(in) == 0 { return }

	for _, tok := range strings.Split (in, sep) {
		ret.PushStr (tok)
	}

	return 
}
