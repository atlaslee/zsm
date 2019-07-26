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
	"errors"
	"github.com/atlaslee/zlog"
	"time"
)

type StateMachineI interface {
	PreLoop() error
	Loop() (bool, error)
	AfterLoop()
	CommandHandle(*Message) (bool, error)
	Run()
	Startup()
	Shutdown()
	State() int
}

type StateMachine struct {
	StateMachineI
	messages chan *Message
	state    int
}

const (
	COMMAND_SHUTDOWN = iota
)

const (
	STATE_INITIALING = iota // New开始执行时的默认状态，到PreLoop完成之前
	STATE_RUNNING           // 具体执行任务的状态
	STATE_STOPPING          // 正在停止的状态。从循环结束开始，到AfterLoop完成
	STATE_FAILED            // 处于失败的状态。
	STATE_STOPPED           // 线程已经停止。
)

var MESSAGE_TYPES []string = []string{"COMMAND_SHUTDOWN"}
var STATES []string = []string{"STATE_INITIALING", "STATE_RUNNING", "STATE_STOPPING", "STATE_FAILED", "STATE_STOPPED"}
var ERR_STARTUP_FAILED = errors.New("Startup failed")

func (this *StateMachine) State() int {
	return this.state
}

func (this *StateMachine) SendMessage(t int) {
	this.SendMessage3(t, nil, nil)
}

func (this *StateMachine) SendMessage2(t int, from interface{}) {
	this.SendMessage3(t, from, nil)
}

func (this *StateMachine) SendMessage3(t int, from, data interface{}) {
	this.messages <- &Message{t, from, data}
}

func (this *StateMachine) Init(Statemachine StateMachineI) {
	zlog.Debugln("SM:", this, "initialing")
	this.StateMachineI = Statemachine
	this.messages = make(chan *Message, 2)
}

func (this *StateMachine) Run() {
	err := this.PreLoop()
	if err != nil {
		this.state = STATE_FAILED
		zlog.Errorln("SM:", this, "failed:", err.Error())
		this.state = STATE_STOPPED
		return
	}
	this.state = STATE_RUNNING
	zlog.Debugln("SM:", this, "running")

	var ok bool
Loop:
	for {
		select {
		case message := <-this.messages:
			switch message.Type {
			case COMMAND_SHUTDOWN:
				zlog.Traceln("MSG:", MESSAGE_TYPES[message.Type], "from", message.From, "received")
				break Loop
			default:
				ok, err = this.CommandHandle(message)
			}
		default:
			ok, err = this.Loop()
		}

		if ok && err == nil {
			continue
		}

		if err != nil {
			this.state = STATE_FAILED
			zlog.Errorln("LOOP:", this, "failed:", err)
		}

		break
	}

	this.state = STATE_STOPPING
	zlog.Traceln("LOOP:", this, "stopping")
	this.AfterLoop()
	this.state = STATE_STOPPED
	zlog.Debugln("LOOP:", this, "stopped")
}

func (this *StateMachine) Startup() {
	zlog.Debugln("SM:", this, "starting")
	go this.Run()
}

func (this *StateMachine) Shutdown() {
	zlog.Debugln("SM:", this, "stopping")

	this.SendMessage2(COMMAND_SHUTDOWN, this)
}

func WaitForStartup(sm StateMachineI) bool {
	for range time.Tick(10 * time.Millisecond) {
		if sm.State() != STATE_INITIALING {
			return true
		}
	}
	return false
}

func WaitForStartupTimeout(sm StateMachineI, dur time.Duration) bool {
	tick := time.Tick(10 * time.Millisecond)
	for {
		select {
		case <-tick:
			if sm.State() != STATE_INITIALING {
				return true
			}
		case <-time.After(dur):
			return false
		}
	}
	return false
}

func WaitForStartupAll(sms []StateMachineI) (ok bool) {
	for range time.Tick(10 * time.Millisecond) {
		ok = false
		for _, sm := range sms {
			if sm.State() == STATE_INITIALING {
				ok = true
				break
			}
		}

		if !ok {
			ok = true
			break
		}
	}
	return
}

func WaitForStartupAllTimeout(sms []StateMachineI, dur time.Duration) (ok bool) {
	tick := time.Tick(10 * time.Millisecond)
Loop:
	for {
		select {
		case <-tick:
			ok = false
			for _, sm := range sms {
				if sm.State() == STATE_INITIALING {
					ok = true
					break
				}
			}

			if !ok {
				ok = true
				break Loop
			}
		case <-time.After(dur):
			ok = false
			break Loop
		}
	}
	return
}

func WaitForShutdown(sm StateMachineI) bool {
	for range time.Tick(10 * time.Millisecond) {
		if sm.State() == STATE_STOPPED {
			return true
		}
	}
	return false
}

func WaitForShutdownTimeout(sm StateMachineI, dur time.Duration) bool {
	tick := time.Tick(10 * time.Millisecond)
	for {
		select {
		case <-tick:
			if sm.State() == STATE_STOPPED {
				return true
			}
		case <-time.After(dur):
			return false
		}
	}
	return false
}

func WaitForShutdownAll(sms []StateMachineI) (ok bool) {
	for range time.Tick(10 * time.Millisecond) {
		ok = false
		for _, sm := range sms {
			if sm.State() == STATE_STOPPED {
				ok = true
				break
			}
		}

		if !ok {
			ok = true
			break
		}
	}
	return
}

func WaitForShutdownAllTimeout(sms []StateMachineI, dur time.Duration) (ok bool) {
	tick := time.Tick(10 * time.Millisecond)
Loop:
	for {
		select {
		case <-tick:
			ok = false
			for _, sm := range sms {
				if sm.State() == STATE_STOPPED {
					ok = true
					break
				}
			}

			if !ok {
				ok = true
				break Loop
			}
		case <-time.After(dur):
			ok = false
			break Loop
		}
	}
	return
}
