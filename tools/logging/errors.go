/** ****************************************************************************************************************** **
	Specifically for logging errors
	
** ****************************************************************************************************************** **/

package logging 

import (
	"github.com/pkg/errors"
	
	"fmt"
	"os"
	"context"
	"log/slog"
	json "github.com/json-iterator/go"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

const (
	ThreadName			= "threadName" // for tracking the original name/function of the calling function
	ThreadId			= "threadId" // for tracking thread ids through the response
)

var (
	ErrReturnToUser  	= errors.New ("Problem with the request")
	ErrKey 				= errors.New ("You need a key")
	ErrPhone			= errors.New ("Please verify phone number")
	ErrFunds  			= errors.New ("You can't afford this")
	ErrNonFatal 		= errors.New ("Non-fatal error")
	ErrUnauthorized		= errors.New ("That's not for you")
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//


//----- Error Handling -----------------------------------------------------------------------------------------------//
type stackTracer interface {
	StackTrace() errors.StackTrace
}

type Logger struct {
}

func (this *Logger) Init () {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions { Level: slog.LevelDebug }))
	slog.SetDefault(logger) // so we just directly slog from somewhere without this original logger
}

func (this *Logger) handleError (ctx context.Context, err error) {
	if err == nil { return } // no error

	var threadName, threadId string
	if ctx != nil {
		threadName, _ = ctx.Value(ThreadName).(string)
		threadId, _ = ctx.Value(ThreadId).(string)
	}
	
	switch errors.Cause (err) {
	case ErrNonFatal, ErrReturnToUser: 
		// we don't need a stack for these types of errors
		slog.Error (err.Error(), 
			slog.String(ThreadName, threadName),
			slog.String(ThreadId, threadId))
	default:
		slog.Error (err.Error(), 
			slog.Any("stacktrace", StackTraceToArray (err)), 
			slog.String(ThreadName, threadName),
			slog.String(ThreadId, threadId))
	}
}

// this handles the context error as well as our resulting error
func (this *Logger) StackTrace (ctx context.Context, err error) {
	// see if we have a context error, do that one first
	if ctx != nil {
		this.handleError (ctx, ctx.Err())
	}
	
	// see if we have another error to include
	this.handleError (ctx, err)
}

// simple wrapper when we want to create a new error and stack trace in the same call
func (this *Logger) TraceErr (ctx context.Context, msg string, params ...interface{}) {
	this.StackTrace (ctx, errors.Errorf(msg, params...))
}

func (this *Logger) DebugObject (ctx context.Context, data interface{}, msg string, params ...interface{}) {
	if data != nil {
		jstr, _ := json.Marshal(data)
		slog.Debug(string(jstr))
	}
	this.TraceErr (ctx, msg, params...)
}

// Checks the context and records an error with the stack if it's bad/expired
func (this *Logger) CtxOk (ctx context.Context) bool {
	this.StackTrace (ctx, errors.WithStack (ctx.Err())) // record any error with a stack
	return ctx.Err() == nil
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// converts an error into an array of strings that contain the stack lines and files
func StackTraceToArray (err error) (ret []string) {
	if err, ok := err.(stackTracer); ok {
		for _, fr := range err.StackTrace() {
			ret = append (ret, fmt.Sprintf ("%+v", fr)) 
			// https://github.com/pkg/errors/blob/5dd12d0cfe7f152f80558d591504ce685299311e/stack.go#L52
			// to reference the sprintf output of fr
		}
	}
	return 
}
