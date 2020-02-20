package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/nhoffmann/monkey/object"

	"github.com/nhoffmann/monkey/evaluator"
	"github.com/nhoffmann/monkey/lexer"
	"github.com/nhoffmann/monkey/parser"
)

// PROMPT denotes the REPL is waiting for input
const PROMPT = ">> "

// Start initializes a REPL
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

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

		evaluated := evaluator.Eval(program, env)

		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}

	}
}

func printParseErrors(out io.Writer, errors []error) {
	for _, error := range errors {
		io.WriteString(out, "Parser Error: ")
		io.WriteString(out, error.Error())
		io.WriteString(out, "\n")
	}
}
