package render

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

var extensions = parser.CommonExtensions | parser.AutoHeadingIDs | parser.Footnotes | parser.SuperSubscript | parser.NoEmptyLineBeforeBlock | parser.DefinitionLists

type meta struct {
	layout   string `yaml:"layout"`
	title    string `yaml:"layout"`
	category string `yaml:"category"`
}

func MdToHTML(loadFrom string) ([]byte, error) {
	md, err := loadFromFile(loadFrom)
	if err != nil {
		return nil, err
	}
	m := meta{}
	body, err := frontmatter.Parse(bytes.NewReader(md), &m)
	if err != nil {
		return nil, err
	}
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(body)
	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.LazyLoadImages | html.TOC | html.FootnoteReturnLinks
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	return markdown.Render(doc, renderer), nil
}

func SaveMdtoHTML(loadFrom, saveTo string) error {
	page, err := MdToHTML(loadFrom)
	if err != nil {
		return err
	}
	if _, found := strings.CutSuffix(saveTo, ".html"); !found {
		return saveToFile(page, fmt.Sprintf("%v.html", saveTo))
	}
	return saveToFile(page, saveTo)
}
