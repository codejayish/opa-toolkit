package format

import (
	"bytes"
	"github.com/open-policy-agent/opa/ast"
	opaformat "github.com/open-policy-agent/opa/format"
)

// Format takes raw Rego source bytes, parses them, and returns formatted output.
func Format(input []byte) ([]byte, error) {
	mod, err := ast.ParseModule("", string(input))
	if err != nil {
		return nil, err
	}

	// 1) Call Ast and capture both values:
	formatted, err := opaformat.Ast(mod)
	if err != nil {
		return nil, err
	}

	// 2) Write the []byte into your buffer:
	buf := bytes.NewBuffer(nil)
	if _, err := buf.Write(formatted); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
