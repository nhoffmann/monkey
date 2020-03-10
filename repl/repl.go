package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/nhoffmann/monkey/object"
	"github.com/nhoffmann/monkey/vm"

	"github.com/nhoffmann/monkey/compiler"

	"github.com/nhoffmann/monkey/lexer"
	"github.com/nhoffmann/monkey/parser"
)

// PROMPT denotes the REPL is waiting for input
const PROMPT = ">> "

// Start initializes a REPL
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		lexer := lexer.NewLexer(line)
		parser := parser.NewParser(lexer)

		program := parser.ParseProgram()

		if len(parser.Errors()) != 0 {
			printParseErrors(out, parser.Errors())
			continue
		}

		compiler := compiler.NewCompilerWithState(symbolTable, constants)
		err := compiler.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Compilation failed: %s\n", err)
			continue
		}

		machine := vm.NewVmWithGlobalsStore(compiler.Bytecode(), globals)
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "Executing bytecode failed: %s\n", err)
		}

		lastPopped := machine.LastPoppedStackElement()
		io.WriteString(out, lastPopped.Inspect())
		io.WriteString(out, "\n")
	}
}

func printParseErrors(out io.Writer, errors []error) {
	for _, error := range errors {
		io.WriteString(out, "Parser Error: ")
		io.WriteString(out, error.Error())
		io.WriteString(out, "\n")
	}
}
