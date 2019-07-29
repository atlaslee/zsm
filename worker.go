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

const (
	CMD_SHUTDOWN = iota
)

const (
	STA_INITIALING = iota // New开始执行时的默认状态，到PreLoop完成之前
	STA_RUNNING           // 具体执行任务的状态
	STA_PAUSING           // 暂停状态
	STA_STOPPING          // 正在停止的状态。从循环结束开始，到AfterLoop完成
	STA_FAILED            // 处于失败的状态。
	STA_STOPPED           // 线程已经停止。
)

var MESSAGE_TYPES []string = []string{"CMD_SHUTDOWN"}
var STATES []string = []string{
	"STA_INITIALING",
	"STA_RUNNING",
	"STA_PAUSING",
	"STA_STOPPING",
	"STA_FAILED",
	"STA_STOPPED"}
var ERR_STARTUP_FAILED = errors.New("Startup failed")

type WorkerI interface {
	PreLoop() error
	AfterLoop()
	CommandHandle(*Message) (bool, error)
	Run()
	Startup()
	Shutdown()
	State() int
}

type Worker struct {
	WorkerI
	messages chan *Message
	state    int
}

func (this *Worker) State() int {
	return this.state
}

func (this *Worker) SendMsg(t int) {
	this.SendMsg3(t, nil, nil)
}

func (this *Worker) SendMsg2(t int, from interface{}) {
	this.SendMsg3(t, from, nil)
}

func (this *Worker) SendMsg3(t int, from, data interface{}) {
	this.messages <- &Message{t, from, data}
}

func (this *Worker) Init(worker WorkerI) {
	zlog.Debugln("WK:", this, "initialing")
	this.WorkerI = worker
	this.messages = make(chan *Message, 2)
}

func (this *Worker) Run() {
	err := this.PreLoop()
	if err != nil {
		this.state = STA_FAILED
		zlog.Errorln("WK:", this, "failed:", err.Error())
		this.state = STA_STOPPED
		return
	}
	this.state = STA_RUNNING
	zlog.Debugln("WK:", this, "running")

	var ok = true
Loop:
	for msg := range this.messages {
		switch msg.Type {
		case CMD_SHUTDOWN:
			zlog.Traceln("WK:", MESSAGE_TYPES[msg.Type], "from", msg.From, "received")
			break Loop
		default:
			ok, err = this.CommandHandle(msg)
		}

		if err != nil {
			this.state = STA_FAILED
			zlog.Errorln("WK:", this, "failed:", err)
			break
		}

		if !ok {
			break
		}
	}

	this.state = STA_STOPPING
	zlog.Traceln("WK:", this, "stopping")
	this.AfterLoop()
	this.state = STA_STOPPED
	zlog.Debugln("WK:", this, "stopped")
}

func (this *Worker) Startup() {
	zlog.Debugln("WK:", this, "starting")
	go this.Run()
}

func (this *Worker) Shutdown() {
	zlog.Debugln("WK:", this, "stopping")

	this.SendMsg2(CMD_SHUTDOWN, this)
}

func WaitForStartup(sm WorkerI) bool {
	for range time.Tick(10 * time.Millisecond) {
		if sm.State() != STA_INITIALING {
			return true
		}
	}
	return false
}

func WaitForStartupTimeout(sm WorkerI, dur time.Duration) bool {
	tick := time.Tick(10 * time.Millisecond)
	timeout := time.Tick(dur)
	for {
		select {
		case <-tick:
			if sm.State() != STA_INITIALING {
				return true
			}
		case <-timeout:
			return false
		}
	}
	return false
}

func WaitForStartupAll(sms []WorkerI) (ok bool) {
	for range time.Tick(10 * time.Millisecond) {
		ok = false
		for _, sm := range sms {
			if sm.State() == STA_INITIALING {
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

func WaitForStartupAllTimeout(sms []WorkerI, dur time.Duration) (ok bool) {
	tick := time.Tick(10 * time.Millisecond)
	timeout := time.Tick(dur)
Loop:
	for {
		select {
		case <-tick:
			ok = false
			for _, sm := range sms {
				if sm.State() == STA_INITIALING {
					ok = true
					break
				}
			}

			if !ok {
				ok = true
				break Loop
			}
		case <-timeout:
			ok = false
			break Loop
		}
	}
	return
}

func WaitForShutdown(sm WorkerI) bool {
	for range time.Tick(10 * time.Millisecond) {
		if sm.State() == STA_STOPPED {
			return true
		}
	}
	return false
}

func WaitForShutdownTimeout(sm WorkerI, dur time.Duration) bool {
	tick := time.Tick(10 * time.Millisecond)
	timeout := time.Tick(dur)
	for {
		select {
		case <-tick:
			if sm.State() == STA_STOPPED {
				return true
			}
		case <-timeout:
			return false
		}
	}
	return false
}

func WaitForShutdownAll(sms []WorkerI) (ok bool) {
	for range time.Tick(10 * time.Millisecond) {
		ok = false
		for _, sm := range sms {
			if sm.State() == STA_STOPPED {
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

func WaitForShutdownAllTimeout(sms []WorkerI, dur time.Duration) (ok bool) {
	tick := time.Tick(10 * time.Millisecond)
	timeout := time.Tick(dur)
Loop:
	for {
		select {
		case <-tick:
			ok = false
			for _, sm := range sms {
				if sm.State() == STA_STOPPED {
					ok = true
					break
				}
			}

			if !ok {
				ok = true
				break Loop
			}
		case <-timeout:
			ok = false
			break Loop
		}
	}
	return
}
