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
)

type MonitorI interface {
	WorkerI
	Loop() (bool, error)
}

type Monitor struct {
	MonitorI
	messages chan *Message
	state    int
}

func (this *Monitor) State() int {
	return this.state
}

func (this *Monitor) SendMsg(t int) {
	this.SendMsg3(t, nil, nil)
}

func (this *Monitor) SendMsg2(t int, from interface{}) {
	this.SendMsg3(t, from, nil)
}

func (this *Monitor) SendMsg3(t int, from, data interface{}) {
	this.messages <- &Message{t, from, data}
}

func (this *Monitor) Init(monitor MonitorI) {
	zlog.Debugln("MON:", this, "initialing")
	this.MonitorI = monitor
	this.messages = make(chan *Message, 2)
}

func (this *Monitor) Run() {
	err := this.PreLoop()
	if err != nil {
		this.state = STA_FAILED
		zlog.Errorln("MON:", this, "failed:", err.Error())
		this.state = STA_STOPPED
		return
	}
	this.state = STA_RUNNING
	zlog.Debugln("MON:", this, "running")

	var ok bool
Loop:
	for {
		select {
		case msg := <-this.messages:
			switch msg.Type {
			case CMD_SHUTDOWN:
				zlog.Traceln("MON:", MESSAGE_TYPES[msg.Type], "from", msg.From, "received")
				break Loop
			default:
				ok, err = this.CommandHandle(msg)
			}
		default:
			ok, err = this.Loop()
		}

		if ok && err == nil {
			continue
		}

		if err != nil {
			this.state = STA_FAILED
			zlog.Errorln("MON:", this, "failed:", err)
		}

		break
	}

	this.state = STA_STOPPING
	zlog.Traceln("MON:", this, "stopping")
	this.AfterLoop()
	this.state = STA_STOPPED
	zlog.Debugln("MON:", this, "stopped")
}

func (this *Monitor) Startup() {
	zlog.Debugln("MON:", this, "starting")
	go this.Run()
}

func (this *Monitor) Shutdown() {
	zlog.Debugln("MON:", this, "stopping")

	this.SendMsg2(CMD_SHUTDOWN, this)
}
