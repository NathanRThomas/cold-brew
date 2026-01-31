/** ****************************************************************************************************************** **
	Flow logic that extends our app object

** ****************************************************************************************************************** **/

package cmd

import (
	"coldbrew/tools"

	"fmt"
	"context"
	"runtime"
	"time"
	"sync"
	json "github.com/json-iterator/go"
	"log/slog"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// generic function prototype we launch using flow control
type flowRunFunc func (context.Context) error 
type flowRunChFunc func (context.Context, interface{}) error 

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// anytime we launch a thread and want to know the stack of where it failed
// call this as a defer
func (this *App) recover () {
	if r := recover(); r != nil {
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		slog.Error (fmt.Sprintf("Recovered from panic: %v", r), slog.Any("stacktrace", buf[:n]))
	}
}

func (this *App) fire (locWg *sync.WaitGroup, fn flowRunFunc, contextTimeout tools.TimeDuration, threadName string, blocking bool) {
	if blocking {
		defer this.recover()

		// create some context for this run
		ctx, cancel := contextTimeout.Context(threadName) // give local context to run
		err := fn(ctx)

		// now check our context, and then error 
		this.StackTrace(ctx, err) // record this error
		
		cancel() // we're done with this local context

	} else {
		// non-blocking

		go func() {
			locWg.Add(1) // add to our sync group
			defer locWg.Done() // make sure we always fire this
			defer this.recover()

			// create some context for this run
			ctx, cancel := contextTimeout.Context(threadName) // give local context to run

			err := fn(ctx)

			// now check our context, and then error 
			this.StackTrace(ctx, err) // record this error
			
			cancel() // we're done with this local context
		}()
	}
}


// background flow "launcher"
// can be used for blocking and non-blocking thread launching
func (this *App) launcher (wg *sync.WaitGroup, fn flowRunFunc, interval, contextTimeout tools.TimeDuration, threadName string, blocking bool) {
	defer wg.Done()

	ticker := interval.Ticker() // how long between calling this function
	done := make(chan bool) // flag so we know when to exit

	locWg := new(sync.WaitGroup) // local sync group for waiting for all the called functions to finish
	locWg.Add(1)

	go func() { // so that each tick enters its own thread
		defer locWg.Done()

		this.fire (locWg, fn, contextTimeout, threadName, blocking) // fire this right away when we start

		for {
			select {
			case <- done: // flag from our channel to indicate this thread should exit
				return 

			case <-ticker.C: // every time the ticker fires, we enter here
				this.fire (locWg, fn, contextTimeout, threadName, blocking)
			}
		}
	}()

	for this.Running { // monitoring thread
		time.Sleep(time.Second)
	}

	ticker.Stop()
	
	done <- true // flag to exit our local thread
	
	locWg.Wait() // wait for our launched threads to finish
}


// this fires a function at an interval. Only runs it in the one thread so this will always finish before firing again
func (this *App) FlowLaunchBlocking (wg *sync.WaitGroup, fn flowRunFunc, interval, contextTimeout tools.TimeDuration, threadName string) {
	this.launcher (wg, fn, interval, contextTimeout, threadName, true)
}

// this fires a function at an interval. 
// this one doesn't block until the previous thread finishes
// which means this can keep launching threads if it takes a long time to finish
func (this *App) FlowLaunch (wg *sync.WaitGroup, fn flowRunFunc, interval, contextTimeout tools.TimeDuration, threadName string) {
	this.launcher (wg, fn, interval, contextTimeout, threadName, false)
}

// wrapper around a flow thread based on channel data
func (this *App) FlowChan (wg *sync.WaitGroup, fn flowRunChFunc, ch chan interface{}, contextTimeout tools.TimeDuration, threadName string) {
	defer wg.Done()

	locWg := new(sync.WaitGroup) // local sync group for waiting for all the called functions to finish
	locWg.Add(1)

	go func() { // so that each tick enters its own thread
		defer locWg.Done()

		for d := range ch {
			if d == nil { return } // nil indicates the channel was closed

			// create some context for this run
			ctx, cancel := contextTimeout.Context(threadName) // give local context to run
			err := fn(ctx, d)

			// now check our context, and then error 
			this.CtxOk (ctx)
			if err != nil {
				this.StackTrace(ctx, err) // record this error

				// to help with debugging, let's include whatever this channel interface was as well
				jstr, _ := json.Marshal(d)
				slog.Debug(fmt.Sprintf("data object: %s\n", string(jstr)))
				time.Sleep(time.Second * 5) // don't hammer these errors if something is wrong
			}
			
			cancel() // we're done with this local context
		}
	}()

	
	locWg.Wait() // wait for the channel be closed, which causes it to exit naturally
}
