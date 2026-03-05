package render

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

type PageInfo struct {
	URL          string
	ModifiedAt   string
	Title        string
	ImgPath      string
	OverviewText string
}

type RenderConfig struct {
	SiteName string
	LogoPath string
	IconPath string
}
type HomeRenderConf struct {
	SiteName string
	LogoPath string
	IconPath string
	Pages    []PageInfo
}

func loadFromFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func saveToFile(data []byte, saveTo string) error {
	dir := filepath.Dir(saveTo)
	if err := os.MkdirAll(dir, 0o755); err != nil || err == os.ErrExist {
		return err
	}
	return os.WriteFile(saveTo, data, 0o644)
}

func getOverviewText(length int, page []byte) string {
	doc, err := html.Parse(bytes.NewReader(page))
	if err != nil {
		return ""
	}
	var result strings.Builder
	var chars int
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "p" {
			text := getText(n)
			for _, r := range text {
				if chars >= 120 {
					break
				}
				result.WriteRune(r)
				chars++
			}
		}
		for c := n.FirstChild; c != nil && chars < 120; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return result.String()
}

func getText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(getText(c))
	}
	return sb.String()
}

func overviewIMG(page []byte) string {
	doc, err := html.Parse(bytes.NewReader(page))
	if err != nil {
		return ""
	}
	var imgSrc string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" && imgSrc == "" {
			for _, attr := range n.Attr {
				if attr.Key == "src" {
					imgSrc = attr.Val
					break
				}
			}
		}
		for c := n.FirstChild; c != nil && imgSrc == ""; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return imgSrc
}
