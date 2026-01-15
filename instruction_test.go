package guac

import (
	"testing"
)

func TestInstruction(t *testing.T) {
	instr := NewInstruction("select", "aaa", "bbb")
	t.Log(instr)
	t.Log(instr.Opcode())
	t.Log(instr.Args())
	for _, e := range instr.Args() {
		t.Log(e.Length(), e.Value())
	}
}
