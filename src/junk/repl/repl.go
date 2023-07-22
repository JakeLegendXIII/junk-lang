package repl

import (
	"bufio"
	"fmt"
	"io"
	"junk/lexer"
	"junk/token"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Printf(PROMPT)        // print prompt
		scanned := scanner.Scan() // scan input
		if !scanned {
			return
		}

		line := scanner.Text()                                                 // get input
		l := lexer.New(line)                                                   // create lexer
		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() { // loop through tokens
			fmt.Printf("%+v\n", tok) // print tokens
		}
	}
}
