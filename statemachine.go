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
	"fmt"
	"github.com/atlaslee/zlog"
	"time"
)

type StateMachineI interface {
	PreLoop() error
	Loop() bool
	AfterLoop()
	CommandHandle(int) bool
	Run()
	Startup() error
	Shutdown()
}

type StateMachine struct {
	StateMachineI
	in, out chan int
}

const (
	COMMAND_SHUT = iota
)

const (
	STATE_READY = iota
	STATE_FAILED
	STATE_CLOSED
)

var COMMANDS []string = []string{"COMMAND_SHUT"}
var STATES []string = []string{"STATE_READY", "STATE_FAILED", "STATE_CLOSED"}
var ERR_STARTUP_FAILED = errors.New("Startup failed.")

func (this *StateMachine) Init(statemachine StateMachineI) {
	this.StateMachineI = statemachine
	this.in = make(chan int, 1)
	this.out = make(chan int, 1)
}

func (this *StateMachine) Run() {
	err := this.PreLoop()
	if err != nil {
		zlog.Errorf("PreLoop failed: %s.\n", err.Error())

		this.out <- STATE_FAILED
		zlog.Traceln("STATE_FAILED sent.")
		return
	}

	this.out <- STATE_READY
	zlog.Traceln("STATE_READY sent.")

Loop:
	for {
		select {
		case command := <-this.in:
			switch command {
			case COMMAND_SHUT:
				zlog.Tracef("%s received.\n", COMMANDS[command])
				break Loop
			default:
				ok := this.CommandHandle(command)
				if !ok {
					zlog.Traceln("Loop stop.")
					break Loop
				}
			}
		default:
			ok := this.Loop()
			if !ok {
				zlog.Traceln("Loop stop.")
				break Loop
			}
			<-time.After(time.Millisecond)
		}
	}

	zlog.Tracef("this.AfterLoop\n")
	this.AfterLoop()

	this.out <- STATE_CLOSED
	zlog.Traceln("STATE_CLOSED sent.")
}

func (this *StateMachine) Startup() (err error) {
	zlog.Debugln("Starting up.")

	go this.Run()

	state := <-this.Out()
	zlog.Tracef("%s received.\n", STATES[state])
	switch state {
	case STATE_READY:
		return
	default:
		zlog.Errorln("Failed to start.")

		return errors.New(fmt.Sprintf("Unexpected state %s received.", STATES[state]))
	}
}

func (this *StateMachine) Shutdown() {
	zlog.Debugln("Shutting down.")

	this.In(COMMAND_SHUT)
	zlog.Traceln("COMMAND_SHUT sent.")

	state := <-this.Out()
	zlog.Tracef("%s received.\n", STATES[state])
	switch state {
	case STATE_CLOSED:
		zlog.Debugln("Closed.")
		return
	default:
		zlog.Debugf("%s received.\n", STATES[state])
		zlog.Warningln("Closed abnormally.")
	}
}

func (this *StateMachine) In(in int) {
	this.in <- in
}

func (this *StateMachine) Out() chan int {
	return this.out
}
