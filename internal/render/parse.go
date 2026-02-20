package render

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

var (
	extensions = parser.CommonExtensions | parser.AutoHeadingIDs | parser.Footnotes | parser.SuperSubscript | parser.NoEmptyLineBeforeBlock | parser.DefinitionLists
	p          = parser.NewWithExtensions(extensions)
)

func MdToHTML(loadFrom string) ([]byte, error) {
	md, err := loadFromFile(loadFrom)
	if err != nil {
		return nil, err
	}
	doc := p.Parse(md)
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
	return saveToFile(page, saveTo)
}
