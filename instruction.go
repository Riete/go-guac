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
	idx := strings.Index(string(e), ".")
	l, _ := strconv.Atoi(string(e)[:idx])
	return l
}

func (e Element) Value() string {
	idx := strings.Index(string(e), ".")
	return string(e)[idx+1:]
}

func NewElement(s string) Element {
	return Element(fmt.Sprintf("%d.%s", len(s), s))
}

// Instruction
// OPCODE,ARG1,ARG2,ARG3,...;
type Instruction string

func (i Instruction) Opcode() Element {
	idx := strings.Index(string(i), ",")
	return Element(string(i)[:idx])
}

func (i Instruction) Args() []Element {
	var elements []Element
	commaIdx := strings.Index(string(i), ",")
	args := string(i)[commaIdx+1:]
	for {
		dotIndex := strings.Index(args, ".")
		length, _ := strconv.Atoi(args[:dotIndex])
		start := dotIndex + 1
		end := start + length
		elements = append(elements, NewElement(args[start:end]))
		if args[end] == ';' {
			break
		}
		args = args[end+1:]
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
