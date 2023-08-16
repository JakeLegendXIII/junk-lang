package repl

import (
	"bufio"
	"fmt"
	"io"
	"junk/lexer"
	"junk/parser"
)

const PROMPT = ">> "
const RACCOON_JUNK = `
  _             _    
  (_)_   _ _ __ | | __
  | | | | | '_ \| |/ /
  | | |_| | | | |   < 
 _/ |\__,_|_| |_|_|\_\
|__/                                                 
`

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Printf(PROMPT)        // print prompt
		scanned := scanner.Scan() // scan input
		if !scanned {
			return
		}

		line := scanner.Text() // get input
		l := lexer.New(line)   // create lexer
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		io.WriteString(out, program.String())
		io.WriteString(out, "\n")
	}
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, RACCOON_JUNK)
	io.WriteString(out, "Woops! We ran into some junk here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
