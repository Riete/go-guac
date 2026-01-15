package guac

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/riete/convert/str"
)

// Element LENGTH.VALUE
type Element string

func (e Element) Length() int {
	l, _ := strconv.Atoi(strings.Split(string(e), ".")[0])
	return l
}

func (e Element) Value() string {
	return strings.Split(string(e), ".")[1]
}

func NewElement(s string) Element {
	return Element(fmt.Sprintf("%d.%s", len(s), s))
}

// Instruction
// OPCODE,ARG1,ARG2,ARG3,...;
type Instruction string

func (i Instruction) Opcode() Element {
	return Element(strings.Split(string(i), ",")[0])
}

func (i Instruction) Args() []Element {
	var elements []Element
	for _, e := range strings.Split(strings.TrimSuffix(string(i), ";"), ",")[1:] {
		elements = append(elements, Element(e))
	}
	return elements
}

func (i Instruction) Byte() []byte {
	return str.ToBytes(string(i))
}

func NewInstruction(opcode string, args ...string) Instruction {
	var elements []string
	for _, i := range append([]string{opcode}, args...) {
		elements = append(elements, string(NewElement(i)))
	}
	return Instruction(strings.Join(elements, ",") + ";")
}
