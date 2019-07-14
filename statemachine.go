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
	CommandHandle(Command int, from, data interface{}) bool
	Run()
	Startup() error
	Shutdown()
}

type StateMachine struct {
	StateMachineI
	Command, State chan *Message
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

func (this *StateMachine) Init(Statemachine StateMachineI) {
	this.StateMachineI = Statemachine
	this.Command = make(chan *Message, 5)
	this.State = make(chan *Message, 1)
}

func (this *StateMachine) SendCommand(Command int) {
	this.SendCommand3(Command, nil, nil)
}

func (this *StateMachine) SendCommand2(Command int, from interface{}) {
	this.SendCommand3(Command, from, nil)
}

func (this *StateMachine) SendCommand3(Command int, from, data interface{}) {
	zlog.Traceln("SendCommand3", Command, from, data)
	this.Command <- MessageNew3(Command, from, data)
}

func (this *StateMachine) ReceiveCommand() (int, interface{}, interface{}) {
	Command := <-this.Command
	return Command.Type, Command.From, Command.Data
}

func (this *StateMachine) SendState(State int) {
	this.SendState3(State, nil, nil)
}

func (this *StateMachine) SendState2(State int, from interface{}) {
	this.SendState3(State, from, nil)
}

func (this *StateMachine) SendState3(State int, from, data interface{}) {
	this.State <- MessageNew3(State, from, data)
}

func (this *StateMachine) ReceiveState() (int, interface{}) {
	State := <-this.State
	return State.Type, State.From
}

func (this *StateMachine) Run() {
	err := this.PreLoop()
	if err != nil {
		zlog.Errorln("PreLoop failed:", err.Error(), ".")

		this.SendState(STATE_FAILED)
		zlog.Traceln("STATE_FAILED sent.")
		return
	}

	this.SendState(STATE_READY)
	zlog.Traceln("STATE_READY sent.")

	tick := time.Tick(10 * time.Millisecond)
Loop:
	for {
		select {
		case Command := <-this.Command:
			switch Command.Type {
			case COMMAND_SHUT:
				zlog.Traceln(COMMANDS[Command.Type], "received.")
				break Loop
			default:
				ok := this.CommandHandle(Command.Type, Command.From, Command.Data)
				if !ok {
					zlog.Traceln("Loop stop.")
					break Loop
				}
			}
		case <-tick:
			ok := this.Loop()
			if !ok {
				zlog.Traceln("Loop stop.")
				break Loop
			}
		}
	}

	zlog.Traceln("this.AfterLoop.")
	this.AfterLoop()

	this.SendState(STATE_CLOSED)
	zlog.Traceln("STATE_CLOSED sent.")
}

func (this *StateMachine) Startup() (err error) {
	zlog.Debugln("Starting up.")

	go this.Run()
	State, _ := this.ReceiveState()
	zlog.Traceln(STATES[State], "received.")

	switch State {
	case STATE_READY:
		return
	default:
		zlog.Errorln("Failed to start.")
		return errors.New(fmt.Sprintf("Unexpected State %s received.", STATES[State]))
	}
}

func (this *StateMachine) Shutdown() {
	zlog.Debugln("Shutting down.")

	this.SendCommand(COMMAND_SHUT)
	zlog.Traceln("COMMAND_SHUT sent.")

	State, _ := this.ReceiveState()
	zlog.Traceln(STATES[State], "received.")
	switch State {
	case STATE_CLOSED:
		zlog.Debugln("Closed.")
		return
	default:
		zlog.Debugln(STATES[State], "received.")
		zlog.Warningln("Closed abnormally.")
	}
}
