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

type MonitorWrapper struct {
	Monitor
	times int
}

func (this *MonitorWrapper) PreLoop() error {
	zlog.Debugln("Starting up.")
	return nil
}

func (this *MonitorWrapper) Loop() (bool, error) {
	<-time.After(1 * time.Second)
	return true, nil
}

func (this *MonitorWrapper) AfterLoop() {
	zlog.Traceln("Runs", this.times, "times totally.")
	zlog.Debugln("Shut down.")
}

func (this *MonitorWrapper) CommandHandle(msg *Message) (bool, error) {
	zlog.Traceln("Command", msg.Type, msg.From, msg.Data, "received.")
	return true, nil
}

func MonitorWrapperNew() (wrapper *MonitorWrapper) {
	wrapper = new(MonitorWrapper)
	wrapper.Init(wrapper)
	return
}

func TestMonitorWrapper(t *testing.T) {
	worker := MonitorWrapperNew()
	worker.Startup()
	WaitForStartup(worker)
	worker.SendMsg(2)
	time.Sleep(2 * time.Second)
	worker.Shutdown()
	WaitForShutdown(worker)
}
