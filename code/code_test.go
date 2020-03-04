package code

import "testing"

func TestMake(t *testing.T) {
	tests := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
		{OpAdd, []int{}, []byte{byte(OpAdd)}},
	}

	for _, test := range tests {
		instruction := Make(test.op, test.operands...)

		if len(instruction) != len(test.expected) {
			t.Errorf(
				"instruction has wrong length. Want %d, got %d.",
				len(test.expected),
				len(instruction),
			)
		}

		for i, b := range test.expected {
			if instruction[i] != test.expected[i] {
				t.Errorf(
					"wrong byte at pos %d. Want %d, got %d",
					i,
					b,
					instruction[i],
				)
			}
		}
	}
}

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
	}

	expected := `0000 OpAdd
0001 OpConstant 2
0004 OpConstant 65535
`

	concatted := Instructions{}

	for _, instruction := range instructions {
		concatted = append(concatted, instruction...)
	}

	if concatted.String() != expected {
		t.Errorf(
			"Instructions wrongly formatted.\nWant %q,\ngot  %q",
			expected,
			concatted.String(),
		)
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op        Opcode
		operands  []int
		bytesRead int
	}{
		{OpConstant, []int{65535}, 2},
	}

	for _, test := range tests {
		instructions := Make(test.op, test.operands...)

		definition, err := Lookup(byte(test.op))
		if err != nil {
			t.Fatalf("definition not found: %q", err)
		}

		operandsRead, n := ReadOperands(definition, instructions[1:])
		if n != test.bytesRead {
			t.Fatalf("number of of operands read is wrong. Want %d, got %d", test.bytesRead, n)
		}

		for i, want := range test.operands {
			if operandsRead[i] != want {
				t.Errorf("Unexpected operand. Want %d, got %d", want, operandsRead[i])
			}
		}
	}
}
