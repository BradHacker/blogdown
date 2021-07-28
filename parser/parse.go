package parser

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type Parser struct {
	md goldmark.Markdown
}

func NewParser() Parser {
	return Parser{
		md: goldmark.New(
			goldmark.WithExtensions(extension.GFM, meta.Meta),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
		),
	}
}

// Parse file takes in a markdown file and parses it into a node tree and corresponding metadata
func (p *Parser) ParseFile(filepath string) (map[string]interface{}, ast.Node, []byte, error) {
	if !strings.HasSuffix(filepath, ".md") {
		return nil, nil, nil, fmt.Errorf("parsing error: input file must be a markdown file with the '.md' extension")
	}
	mdFileContent, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parsing error: %v", err)
	}

	metaContext := parser.NewContext()
	documentNode := p.md.Parser().Parse(text.NewReader(mdFileContent), parser.WithContext(metaContext))

	metadata, err := meta.TryGet(metaContext)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parsing error: error getting metadata: %v", err)
	}
	return metadata, documentNode, mdFileContent, nil
}
