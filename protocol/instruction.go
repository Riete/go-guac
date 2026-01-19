package protocol

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/riete/convert/str"
)

// Element LENGTH.VALUE
// Each element of the list has a positive decimal integer length prefix separated by the value of the element by a period.
// This length denotes the number of Unicode characters in the value of the element, which is encoded in UTF-8
// https://guacamole.apache.org/doc/gug/guacamole-protocol.html
type Element string

func (e Element) Length() int {
	s := string(e)
	idx := strings.Index(s, ".")
	l, _ := strconv.Atoi(s[:idx])
	return l
}

func (e Element) Value() string {
	s := string(e)
	idx := strings.Index(s, ".")
	return s[idx+1:]
}

func NewElement(s string) Element {
	length := utf8.RuneCountInString(s)
	return Element(strconv.Itoa(length) + "." + s)
}

// Instruction
// OPCODE,ARG1,ARG2,ARG3,...;
// Each instruction is a comma-delimited list followed by a terminating semicolon,
// where the first element of the list is the instruction opcode,
// and all following elements are the arguments for that instruction
// https://guacamole.apache.org/doc/gug/guacamole-protocol.html#design
type Instruction string

func (i Instruction) Opcode() Element {
	s := string(i)
	idx := strings.Index(s, ",")
	if idx == -1 {
		return Element(s[:strings.LastIndex(s, ";")])
	}
	return Element(s[:idx])
}

func (i Instruction) Args() []Element {
	s := string(i)
	commaIdx := strings.Index(s, ",")
	if commaIdx == -1 {
		return []Element{}
	}
	args := s[commaIdx+1:]
	var elements []Element
	for {
		dotIndex := strings.Index(args, ".")
		length, _ := strconv.Atoi(args[:dotIndex])
		start := dotIndex + 1
		runeCount := 0
		end := start
		for runeCount < length {
			_, size := utf8.DecodeRuneInString(args[end:])
			end += size
			runeCount++
		}
		elements = append(elements, Element(args[:end]))
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
	for _, arg := range append([]string{opcode}, args...) {
		elements = append(elements, string(NewElement(arg)))
	}
	return Instruction(strings.Join(elements, ",") + ";")
}
