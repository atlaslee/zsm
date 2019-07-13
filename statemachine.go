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
	CommandHandle(command int, from, data interface{}) bool
	Run()
	Startup() error
	Shutdown()
}

type StateMachine struct {
	StateMachineI
	command, state chan *Message
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
	this.command = make(chan *Message, 5)
	this.state = make(chan *Message, 1)
}

func (this *StateMachine) SendCommand(command int) {
	this.SendCommand3(command, nil, nil)
}

func (this *StateMachine) SendCommand2(command int, from interface{}) {
	this.SendCommand3(command, from, nil)
}

func (this *StateMachine) SendCommand3(command int, from, data interface{}) {
	zlog.Traceln("SendCommand3", command, from, data)
	this.command <- MessageNew3(command, from, data)
}

func (this *StateMachine) ReceiveCommand() (int, interface{}, interface{}) {
	command := <-this.command
	return command.Type, command.From, command.Data
}

func (this *StateMachine) SendState(state int) {
	this.SendState3(state, nil, nil)
}

func (this *StateMachine) SendState2(state int, from interface{}) {
	this.SendState3(state, from, nil)
}

func (this *StateMachine) SendState3(state int, from, data interface{}) {
	this.state <- MessageNew3(state, from, data)
}

func (this *StateMachine) ReceiveState() (int, interface{}) {
	state := <-this.state
	return state.Type, state.From
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
		case command := <-this.command:
			switch command.Type {
			case COMMAND_SHUT:
				zlog.Traceln(COMMANDS[command.Type], "received.")
				break Loop
			default:
				ok := this.CommandHandle(command.Type, command.From, command.Data)
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
	state, _ := this.ReceiveState()
	zlog.Traceln(STATES[state], "received.")

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

	this.SendCommand(COMMAND_SHUT)
	zlog.Traceln("COMMAND_SHUT sent.")

	state, _ := this.ReceiveState()
	zlog.Traceln(STATES[state], "received.")
	switch state {
	case STATE_CLOSED:
		zlog.Debugln("Closed.")
		return
	default:
		zlog.Debugln(STATES[state], "received.")
		zlog.Warningln("Closed abnormally.")
	}
}
