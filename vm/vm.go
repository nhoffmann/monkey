package vm

import (
	"fmt"

	"github.com/nhoffmann/monkey/code"
	"github.com/nhoffmann/monkey/compiler"
	"github.com/nhoffmann/monkey/object"
)

const StackSize = 2048

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}

type VM struct {
	constants    []object.Object
	instructions code.Instructions
	stack        []object.Object
	stackPointer int
}

func NewVm(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		stackPointer: 0,
	}
}

func (vm *VM) Run() error {
	for insPointer := 0; insPointer < len(vm.instructions); insPointer++ {
		// fetch
		op := code.Opcode(vm.instructions[insPointer])

		// decode
		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[insPointer+1:])
			insPointer += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSubtract, code.OpMultiply, code.OpDivide:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpGreaterThan, code.OpEqual, code.OpNotEqual:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vm *VM) LastPoppedStackElement() object.Object {
	return vm.stack[vm.stackPointer]
}

func (vm *VM) push(obj object.Object) error {
	if vm.stackPointer >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.stackPointer] = obj
	vm.stackPointer++

	return nil
}

func (vm *VM) pop() object.Object {
	obj := vm.stack[vm.stackPointer-1]
	vm.stackPointer--

	return obj
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	if leftType == object.INTEGER && rightType == object.INTEGER {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER || right.Type() == object.INTEGER {
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf(
			"unknown operator: %d (%s %s)",
			op,
			left.Type(),
			right.Type(),
		)
	}
}

func (vm *VM) executeIntegerComparison(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue == leftValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue != leftValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result int64

	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSubtract:
		result = leftValue - rightValue
	case code.OpMultiply:
		result = leftValue * rightValue
	case code.OpDivide:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknonw integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

func nativeBoolToBooleanObject(input bool) object.Object {
	if input {
		return True
	}

	return False
}
