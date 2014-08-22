package ujihisa

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

type value interface{}

func toInt(v value) int {
	if v, ok := v.(int); ok {
		return v
	}
	return 0
}

func toChar(v value) string {
	if v, ok := v.(int); ok {
		return string(rune(v))
	}
	if v, ok := v.(rune); ok {
		return string(rune(v))
	}
	return ""
}

func toString(v value) string {
	if v, ok := v.(string); ok {
		return v
	}
	if v, ok := v.(int); ok {
		return fmt.Sprint(v)
	}
	return ""
}

type VM struct {
	done   bool
	code   []byte
	pc     int
	result value
	Debug  bool
}

func NewVM() *VM {
	return &VM{
		done:   false,
		code:   nil,
		pc:     0,
		result: nil,
		Debug:  false,
	}
}

func (vm *VM) next() byte {
	c := vm.code[vm.pc]
	vm.pc++
	return c
}

func (vm *VM) restart() {
	vm.pc = 0
}

func (vm *VM) reset() {
	vm.done = false
}

func (vm *VM) parse(s string) error {
	var bc []byte

	cmdA := "便利"
	cmdB := "感極まってきました"
	cmdN := "かなり"

	for len(s) > 0 {
		switch {
		case strings.HasPrefix(s, cmdA):
			bc = append(bc, 'A')
			s = s[len(cmdA):]
		case strings.HasPrefix(s, cmdB):
			bc = append(bc, 'B')
			s = s[len(cmdB):]
		case strings.HasPrefix(s, cmdN):
			bc = append(bc, 'N')
			s = s[len(cmdN):]
		case strings.HasPrefix(s, "\n"):
			bc = append(bc, 'C')
			s = s[1:]
		default:
			s = s[1:]
		}
	}
	vm.code = bc
	return nil
}

var bb int

func (vm *VM) do(s string) bool {
	if vm.done {
		return false
	}

	var isLabel, isNumber bool

	if strings.HasSuffix(s, "l") {
		isLabel = true
		s = s[:len(s)-1]
	}
	if strings.HasSuffix(s, "n") {
		isNumber = true
		s = s[:len(s)-1]
	}

	code := string(vm.code[vm.pc : vm.pc+len(s)])
	if code != s {
		return false
	}
	vm.pc += len(s)

	if isNumber {
		sign := -1
		if vm.next() == 'A' {
			sign = 1
		}
		num := 0
		for {
			c := vm.next()
			if c == 'C' {
				break
			}
			num *= 2
			if c == 'B' {
				num++
			}
		}
		vm.result = sign * num
	} else if isLabel {
		var label []byte
		for {
			c := vm.next()
			if c == 'C' {
				break
			}
			label = append(label, c)
		}
		vm.result = string(label)
	}
	vm.done = true
	return true
}

func (vm *VM) Run(s string) error {
	var err error

	err = vm.parse(s)
	if err != nil {
		return err
	}

	if vm.Debug {
		fmt.Println("command list:", string(vm.code))
	}

	stack := []value{}
	pcstack := []int{}
	heap := map[int]value{}
	last := -1

	cmds := []string{
		"AAn", "ACA", "ACB", "ACC",
		"BAAA", "BAAB", "BAAC", "BABA", "BABB",
		"BBA", "BBB",
		"CAAl", "CABl", "CACl", "CBAl", "CBBl", "CBC", "CCC",
		"BCAA", "BCAB", "BCBA", "BCBB",
		"N",
	}
	labels := map[string]int{}

	for vm.pc < len(vm.code) {
		if last == vm.pc {
			if vm.pc != len(vm.code) - 1 || vm.code[vm.pc] != 'C' {
				return errors.New("Seems broken byte codes")
			}
			return nil
		}
		last = vm.pc
		vm.reset()
		for _, cmd := range cmds {
			pc := vm.pc
			if vm.do(cmd) {
				if cmd == "CAAl" {
					labels[toString(vm.result)] = vm.pc
				}
				if vm.Debug {
					fmt.Fprintf(os.Stderr, "position: %d command: %s\n", pc, cmd)
				}
			}
		}
	}
	vm.restart()

	last = -1
	for vm.pc < len(vm.code) {
		if last == vm.pc {
			return errors.New("Seems broken byte codes")
		}
		last = vm.pc

		if vm.Debug {
			fmt.Fprintf(os.Stderr, "position=%d\n", vm.pc)
			fmt.Fprintf(os.Stderr, "stack=%s\n", fmt.Sprint(stack))
		}

		vm.reset()
		if vm.do("AAn") {
			stack = append(stack, vm.result)
		}
		if vm.do("ACA") {
			stack = append(stack, stack[len(stack)-1])
		}
		if vm.do("ACB") {
			stack[len(stack)-2], stack[len(stack)-1] = stack[len(stack)-1], stack[len(stack)-2]
		}
		if vm.do("ACC") {
			stack = stack[:len(stack)-1]
		}
		if vm.do("BA") {
			left, right := stack[len(stack)-2], stack[len(stack)-1]
			stack = stack[:len(stack)-2]
			vm.reset()
			if vm.do("AA") {
				stack = append(stack, toInt(left)+toInt(right))
			}
			if vm.do("AB") {
				stack = append(stack, toInt(left)-toInt(right))
			}
			if vm.do("AC") {
				stack = append(stack, toInt(left)*toInt(right))
			}
			if vm.do("BA") {
				stack = append(stack, toInt(left)/toInt(right))
			}
			if vm.do("BB") {
				stack = append(stack, toInt(left)%toInt(right))
			}
			continue
		}
		if vm.do("BBA") {
			index, value := toInt(stack[len(stack)-2]), stack[len(stack)-1]
			stack = stack[:len(stack)-2]
			heap[index] = value
		}
		if vm.do("BBB") {
			index := toInt(stack[len(stack)-1])
			stack = stack[:len(stack)-1]
			stack = append(stack, heap[index])
		}
		vm.do("CAAl")
		if vm.do("CABl") {
			pcstack = append(pcstack, vm.pc)
			vm.pc = labels[toString(vm.result)]
		}
		if vm.do("CACl") {
			vm.pc = labels[toString(vm.result)]
		}
		if vm.do("CBAl") {
			v := toInt(stack[len(stack)-1])
			stack = stack[:len(stack)-1]
			if v == 0 {
				vm.pc = labels[toString(vm.result)]
			}
		}
		if vm.do("CBBl") {
			v := toInt(stack[len(stack)-1])
			stack = stack[:len(stack)-1]
			if v < 0 {
				vm.pc = labels[toString(vm.result)]
			}
		}
		if vm.do("CBC") {
			addr := pcstack[len(pcstack)-1]
			pcstack = pcstack[:len(pcstack)-1]
			vm.pc = addr
		}
		if vm.do("CCC") {
			return nil
		}
		if vm.do("BCAA") {
			c := toChar(stack[len(stack)-1])
			stack = stack[:len(stack)-1]
			fmt.Print(c)
		}
		if vm.do("BCAB") {
			n := stack[len(stack)-1].(int)
			stack = stack[:len(stack)-1]
			fmt.Print(n)
		}
		if vm.do("BCBA") {
			r, _, err := bufio.NewReader(os.Stdin).ReadRune()
			if err != nil {
				return err
			}
			addr := stack[len(stack)-1].(int)
			stack = stack[:len(stack)-1]
			heap[addr] = string(rune(r))
		}
		if vm.do("BCBB") {
			var n int
			_, err := fmt.Scan(&n)
			if err != nil {
				return err
			}
			addr := stack[len(stack)-1].(int)
			stack = stack[:len(stack)-1]
			heap[addr] = n
		}
		if vm.do("N") {
			fmt.Println("かなり")
		}
	}
	return nil
}
