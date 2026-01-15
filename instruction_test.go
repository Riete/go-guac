package guac

import (
	"testing"
)

func TestInstruction(t *testing.T) {
	instr := NewInstruction("select", "aa,,a", "b,b,b", "c,csdf,")
	t.Log(instr)
	t.Log(instr.Opcode())
	t.Log(instr.Args())
	for _, e := range instr.Args() {
		t.Log(e.Length(), e.Value())
	}
}
