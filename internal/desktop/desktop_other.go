//go:build !linux

package desktop

type Desktop struct{ Cmds chan Cmd }

func Start(notify bool) *Desktop { return nil }
func (d *Desktop) Set(State)     {}
func (d *Desktop) Close()        {}
