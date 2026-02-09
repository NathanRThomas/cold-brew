
package zerobounce

import (
	"github.com/stretchr/testify/assert"

	"coldbrew/tools"
	
	"testing"
	"context"
	"time"
)

var cfg struct {
	ZeroBounce string
}

func TestNetValidateEmail1a (t *testing.T) {

	err := tools.TestingLoadConfig (&cfg)
	tools.TestingStackTrace(t, err)

	ctx, cancel := context.WithTimeout (context.Background(), time.Minute)
	defer cancel()

	ok, typo, err := ValidateEmail (ctx, cfg.ZeroBounce, "example@gmail.com")
	tools.TestingStackTrace(t, err)

	assert.Equal (t, true, ok)
	assert.Equal (t, "", typo)
}

func TestNetValidateEmail2a (t *testing.T) {

	err := tools.TestingLoadConfig (&cfg)
	tools.TestingStackTrace(t, err)

	ctx, cancel := context.WithTimeout (context.Background(), time.Minute)
	defer cancel()

	ok, typo, err := ValidateEmail (ctx, cfg.ZeroBounce, "example@gmil.com")
	tools.TestingStackTrace(t, err)

	assert.Equal (t, false, ok)
	assert.Equal (t, "example@gmail.com", typo)
}

