package controller

import (
	"github.com/drpaneas/n64/debug"
	"github.com/drpaneas/n64/rcp/serial"
	"github.com/drpaneas/n64/rcp/serial/joybus"
)

var (
	cmdAllStates      *serial.CommandBlock
	cmdAllStatesPorts [4]joybus.ControllerStateCommand

	cmdAllInfo      *serial.CommandBlock
	cmdAllInfoPorts [4]joybus.InfoCommand
)

func init() {
	var err error

	cmdAllStates = serial.NewCommandBlock(serial.CmdConfigureJoybus)
	for i := range cmdAllStatesPorts {
		cmdAllStatesPorts[i], err = joybus.NewControllerStateCommand(cmdAllStates)
		debug.AssertErrNil(err)
	}
	err = joybus.ControlByte(cmdAllStates, joybus.CtrlAbort)
	debug.AssertErrNil(err)

	cmdAllInfo = serial.NewCommandBlock(serial.CmdConfigureJoybus)
	for i := range cmdAllInfoPorts {
		cmdAllInfoPorts[i], err = joybus.NewInfoCommand(cmdAllInfo)
		debug.AssertErrNil(err)
	}
	err = joybus.ControlByte(cmdAllInfo, joybus.CtrlAbort)
	debug.AssertErrNil(err)

	for i := range States {
		States[i].Port.number = uint8(i + 1)
	}
}

type allControllers [4]Controller

var States allControllers

func (p *allControllers) Poll() {
	// poll info
	for _, cmd := range cmdAllInfoPorts {
		cmd.Reset()
	}
	serial.Run(cmdAllInfo)

	// poll states
	for _, cmd := range cmdAllStatesPorts {
		cmd.Reset()
	}

	serial.Run(cmdAllStates)

	for i := range p {
		var err error

		p[i].Port.last = p[i].Port.current
		dev, flags, err := cmdAllInfoPorts[i].Info()
		p[i].Port.current.device = dev
		p[i].Port.current.flags = flags
		p[i].Port.err = err

		p[i].last = p[i].current
		cur := &p[i].current
		cur.down, cur.xAxis, cur.yAxis, p[i].err = cmdAllStatesPorts[i].State()
	}
}
