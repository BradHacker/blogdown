package renderer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"text/template"

	"github.com/yuin/goldmark/ast"
)

const (
	// Block Templates
	DocumentTemplatePath        = "../templates/block/document.tmpl.html"
	TextBlockTemplatePath       = "../templates/block/text-block.tmpl.html"
	ParagraphTemplatePath       = "../templates/block/paragraph.tmpl.html"
	HeadingTemplatePath         = "../templates/block/heading.tmpl.html"
	ThematicBreakTemplatePath   = "../templates/block/thematic-break.tmpl.html"
	CodeBlockTemplatePath       = "../templates/block/code-block.tmpl.html"
	FencedCodeBlockTemplatePath = "../templates/block/fenced-code-block.tmpl.html"
	BlockquoteTemplatePath      = "../templates/block/blockquote.tmpl.html"
	ListTemplatePath            = "../templates/block/list.tmpl.html"
	ListItemTemplatePath        = "../templates/block/list-item.tmpl.html"
	HTMLBlockTemplatePath       = "../templates/block/html-block.tmpl.html"

	// Inline Templates
	TextTemplatePath     = "../templates/inline/text.tmpl.html"
	StringTemplatePath   = "../templates/inline/string.tmpl.html"
	CodeSpanTemplatePath = "../templates/inline/code-span.tmpl.html"
	EmphasisTemplatePath = "../templates/inline/emphasis.tmpl.html"
	LinkTemplatePath     = "../templates/inline/link.tmpl.html"
	ImageTemplatePath    = "../templates/inline/image.tmpl.html"
	AutoLinkTemplatePath = "../templates/inline/auto-link.tmpl.html"
	RawHTMLTemplatePath  = "../templates/inline/raw-html.tmpl.html"
)

var (
	// Map Templates to AST Types
	KindTemplateMap map[ast.NodeKind]string
)

type PageMeta struct {
	Title       string
	Description string
	Slug        string
	Path        string
}

type TemplateData struct {
	Meta     PageMeta
	Config   map[string]interface{}
	Content  string
	Children string
}

func InitKindTemplateMap() {
	rendererPath := "./"
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		rendererPath = path.Dir(filename)
	}
	KindTemplateMap = make(map[ast.NodeKind]string)
	KindTemplateMap[ast.KindDocument] = path.Join(rendererPath, DocumentTemplatePath)
	KindTemplateMap[ast.KindTextBlock] = path.Join(rendererPath, TextBlockTemplatePath)
	KindTemplateMap[ast.KindParagraph] = path.Join(rendererPath, ParagraphTemplatePath)
	KindTemplateMap[ast.KindHeading] = path.Join(rendererPath, HeadingTemplatePath)
	KindTemplateMap[ast.KindThematicBreak] = path.Join(rendererPath, ThematicBreakTemplatePath)
	KindTemplateMap[ast.KindCodeBlock] = path.Join(rendererPath, CodeBlockTemplatePath)
	KindTemplateMap[ast.KindFencedCodeBlock] = path.Join(rendererPath, FencedCodeBlockTemplatePath)
	KindTemplateMap[ast.KindBlockquote] = path.Join(rendererPath, BlockquoteTemplatePath)
	KindTemplateMap[ast.KindList] = path.Join(rendererPath, ListTemplatePath)
	KindTemplateMap[ast.KindListItem] = path.Join(rendererPath, ListItemTemplatePath)
	KindTemplateMap[ast.KindHTMLBlock] = path.Join(rendererPath, HTMLBlockTemplatePath)
	KindTemplateMap[ast.KindText] = path.Join(rendererPath, TextTemplatePath)
	KindTemplateMap[ast.KindString] = path.Join(rendererPath, StringTemplatePath)
	KindTemplateMap[ast.KindCodeSpan] = path.Join(rendererPath, CodeSpanTemplatePath)
	KindTemplateMap[ast.KindEmphasis] = path.Join(rendererPath, EmphasisTemplatePath)
	KindTemplateMap[ast.KindLink] = path.Join(rendererPath, LinkTemplatePath)
	KindTemplateMap[ast.KindImage] = path.Join(rendererPath, ImageTemplatePath)
	KindTemplateMap[ast.KindAutoLink] = path.Join(rendererPath, AutoLinkTemplatePath)
	KindTemplateMap[ast.KindRawHTML] = path.Join(rendererPath, RawHTMLTemplatePath)
}

func RenderAst(pageMetadata map[string]interface{}, documentNode ast.Node, inputFileBytes []byte) error {
	pageTitleInt, exists := pageMetadata["title"]
	if !exists {
		return fmt.Errorf("rendering error: page does not contain \"title\" in metadata")
	}
	pageTitle := fmt.Sprintf("%s", pageTitleInt)
	pageDescriptionInt, exists := pageMetadata["description"]
	if !exists {
		return fmt.Errorf("rendering error: page does not contain \"description\" in metadata")
	}
	pageDescription := fmt.Sprintf("%s", pageDescriptionInt)
	pageSlugInt, exists := pageMetadata["slug"]
	if !exists {
		return fmt.Errorf("rendering error: page does not contain \"slug\" in metadata")
	}
	pageSlug := fmt.Sprintf("%s", pageSlugInt)
	pagePathInt, exists := pageMetadata["path"]
	if !exists {
		return fmt.Errorf("rendering error: page does not contain \"path\" in metadata")
	}
	pagePath := fmt.Sprintf("%s", pagePathInt)
	if !strings.HasPrefix(pagePath, "/") {
		return fmt.Errorf("rendering error: page path doesn't being with a \"/\" (forward slash)")
	}

	pageMeta := PageMeta{
		Title:       pageTitle,
		Description: pageDescription,
		Slug:        pageSlug,
		Path:        DocumentTemplatePath,
	}
	output, err := renderAstNode(pageMeta, documentNode, inputFileBytes)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path.Join("build", pagePath)); os.IsNotExist(err) {
		err := os.MkdirAll(path.Join("build", pagePath), os.ModeDir)
		if err != nil {
			return err
		}
	}

	err = ioutil.WriteFile(path.Join("build", pagePath, "/index.html"), []byte(output), os.ModeAppend)
	if err != nil {
		return err
	}
	return nil
}

func renderAstNode(pageMetadata PageMeta, n ast.Node, inputFileBytes []byte) (output string, err error) {
	firstChildContent := ""
	if n.HasChildren() {
		firstChildContent, err = renderAstNode(pageMetadata, n.FirstChild(), inputFileBytes)
		if err != nil {
			return
		}
	}
	templateContent, err := getTemplateContent(n.Kind())
	if err != nil {
		return
	}
	templateConfig := generateTemplateDataConfig(n)
	output, err = renderTemplateToString(templateContent, TemplateData{
		Meta:     pageMetadata,
		Config:   templateConfig,
		Content:  generateTemplateDataContent(n, inputFileBytes),
		Children: firstChildContent,
	})
	if err != nil {
		return
	}
	if n.NextSibling() != nil {
		nextSiblingContent, err := renderAstNode(pageMetadata, n.NextSibling(), inputFileBytes)
		if err != nil {
			return "", err
		}
		output = fmt.Sprintf("%s\n%s", output, nextSiblingContent)
	}
	return
}

func generateTemplateDataContent(n ast.Node, inputFileBytes []byte) string {
	switch n.Kind() {
	case ast.KindText:
		text := n.(*ast.Text)
		content := string(text.Segment.Value(inputFileBytes))
		if text.SoftLineBreak() {
			content = fmt.Sprintf("%s\n", content)
		} else {
			content = fmt.Sprintf("%s<br/>\n", content)
		}
		return content
	case ast.KindFencedCodeBlock:
		fencedCodeBlock := n.(*ast.FencedCodeBlock)
		content := ""
		l := fencedCodeBlock.Lines().Len()
		for i := 0; i < l; i++ {
			line := fencedCodeBlock.Lines().At(i)
			content = fmt.Sprintf("%s%s", content, line.Value(inputFileBytes))
		}
		// TODO: Add syntax highlighting somehow
		return content
	default:
		return ""
	}
}

func generateTemplateDataConfig(n ast.Node) map[string]interface{} {
	config := make(map[string]interface{})
	switch n.Kind() {
	case ast.KindHeading:
		config["level"] = n.(*ast.Heading).Level
	case ast.KindEmphasis:
		if n.(*ast.Emphasis).Level == 1 {
			config["tagType"] = "em"
		} else {
			config["tagType"] = "strong"
		}
	}
	return config
}

func getTemplateContent(nodeKind ast.NodeKind) (string, error) {
	if len(KindTemplateMap) == 0 {
		InitKindTemplateMap()
	}
	templatePath, exists := KindTemplateMap[nodeKind]
	if !exists {
		return "", fmt.Errorf("rendering error: node kind (%s) doesn't have a template assigned to it", nodeKind)
	}
	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("rendering error: while reading template %s: %v", templatePath, err)
	}
	return string(templateContent), nil
}

func renderTemplateToString(templateContent string, templateData TemplateData) (string, error) {
	t, err := template.New("document").Parse(templateContent)
	if err != nil {
		return "", err
	}
	var outputBuffer bytes.Buffer
	err = t.Execute(&outputBuffer, templateData)
	if err != nil {
		return "", err
	}
	return outputBuffer.String(), nil
}
