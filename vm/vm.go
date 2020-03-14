package vm

import (
	"fmt"

	"github.com/nhoffmann/monkey/code"
	"github.com/nhoffmann/monkey/compiler"
	"github.com/nhoffmann/monkey/object"
)

const StackSize = 2048
const GlobalsSize = 65536

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

type VM struct {
	constants    []object.Object
	instructions code.Instructions
	stack        []object.Object
	stackPointer int
	globals      []object.Object
}

func NewVm(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		stackPointer: 0,
		globals:      make([]object.Object, GlobalsSize),
	}
}

func NewVmWithGlobalsStore(bytecode *compiler.Bytecode, globals []object.Object) *VM {
	vm := NewVm(bytecode)
	vm.globals = globals
	return vm
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
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}
		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}
		case code.OpJump:
			position := int(code.ReadUint16(vm.instructions[insPointer+1:]))
			insPointer = position - 1
		case code.OpJumpNotTruthy:
			position := int(code.ReadUint16(vm.instructions[insPointer+1:]))

			insPointer += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				insPointer = position - 1
			}
		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[insPointer+1:])
			insPointer += 2
			vm.globals[globalIndex] = vm.pop()
		case code.OpGetGlobal:
			globalIndex := code.ReadUint16((vm.instructions[insPointer+1:]))
			insPointer += 2
			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		case code.OpArray:
			numElements := int(code.ReadUint16(vm.instructions[insPointer+1:]))
			insPointer += 2

			array := vm.buildArray(vm.stackPointer-numElements, vm.stackPointer)
			vm.stackPointer = vm.stackPointer - numElements

			err := vm.push(array)
			if err != nil {
				return err
			}
		case code.OpHash:
			numElements := int(code.ReadUint16(vm.instructions[insPointer+1:]))
			insPointer += 2

			hash, err := vm.buildHash(vm.stackPointer-numElements, vm.stackPointer)
			if err != nil {
				return err
			}
			vm.stackPointer = vm.stackPointer - numElements

			err = vm.push(hash)
			if err != nil {
				return err
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
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

	switch {
	case leftType == object.INTEGER && rightType == object.INTEGER:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.STRING && rightType == object.STRING:
		return vm.executeBinaryStringOperation(op, left, right)
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

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
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

func (vm *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}

	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(startIndex, endIndex int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: value}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hashkey: %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: hashedPairs}, nil
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY && index.Type() == object.INTEGER:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(left, index object.Object) error {
	arrayObject := left.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if i < 0 || i > max {
		return vm.push(Null)
	}

	return vm.push(arrayObject.Elements[i])
}

func (vm *VM) executeHashIndex(left, index object.Object) error {
	hashObject := left.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

func (vm *VM) executeBinaryStringOperation(op code.Opcode, left, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	return vm.push(&object.String{Value: leftValue + rightValue})
}

func nativeBoolToBooleanObject(input bool) object.Object {
	if input {
		return True
	}

	return False
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}

}
