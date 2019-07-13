/* The MIT License (MIT)
Copyright © 2018 by Atlas Lee(atlas@fpay.io)

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the “Software”),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
DEALINGS IN THE SOFTWARE.
*/

package zsm

import (
	"github.com/atlaslee/zlog"
	"testing"
	"time"
)

type StateMachineWrapper struct {
	StateMachine
	times int
}

func (this *StateMachineWrapper) PreLoop() error {
	zlog.Debugln("Starting up.")
	return nil
}

func (this *StateMachineWrapper) Loop() bool {
	this.times++
	time.Sleep(10 * time.Millisecond)
	return true
}

func (this *StateMachineWrapper) AfterLoop() {
	zlog.Traceln("Runs", this.times, "times totally.")
	zlog.Debugln("Shut down.")
}

func (this *StateMachineWrapper) CommandHandle(command int, from, data interface{}) bool {
	zlog.Traceln("Command", command, from, data, "received.")
	return true
}

func StateMachineWrapperNew() (wrapper *StateMachineWrapper) {
	wrapper = new(StateMachineWrapper)
	wrapper.Init(wrapper)
	return
}

func TestStateMachine(t *testing.T) {
	statemachine := StateMachineWrapperNew()
	statemachine.Startup()
	statemachine.SendCommand(2)
	time.Sleep(2 * time.Second)
	statemachine.Shutdown()
}
