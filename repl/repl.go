package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/nhoffmann/monkey/lexer"
	"github.com/nhoffmann/monkey/token"
)

// PROMPT denotes the REPL is waiting for input
const PROMPT = ">> "

// Start initializes a REPL
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		lexer := lexer.NewLexer(line)

		for tok := lexer.NextToken(); tok.Type != token.EOF; tok = lexer.NextToken() {
			fmt.Fprintf(out, "%+v\n", tok)
		}
	}
}
