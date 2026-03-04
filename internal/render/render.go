// Package render is about generating full html pages from templates and parsing markdown to html.
package render

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	paths "cms/internal"
)

// assets/templates

func RenderTemplates(base string, data any, tmpls []string) ([]byte, error) {
	tmpl, err := template.ParseFiles(tmpls...)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	err = tmpl.ExecuteTemplate(&buffer, base, data)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

type dataStruct struct {
	Title        string
	Content      template.HTML
	Style        string
	Script       string
	SiteName     string
	Year         int
	SideBarLinks []Link
}

func SaveMdtoHTML(loadFrom, saveTo string, rndrConf *RenderConfig) error {
	page, title, err := parseMdToHTML(loadFrom)
	if err != nil {
		return err
	}
	title += fmt.Sprintf(" | %v", rndrConf.SiteName)
	fileName, _ := strings.CutSuffix(filepath.Base(loadFrom), ".md")
	entries, err := os.ReadDir(filepath.Join(paths.AssetsPath, "templates"))
	if err != nil {
		return err
	}
	templates := []string{}
	for _, e := range entries {
		_, has := strings.CutSuffix(e.Name(), ".tmpl")
		if has && !e.IsDir() {
			templates = append(templates, filepath.Join(paths.AssetsPath, "templates", e.Name()))
		}
	}

	data := dataStruct{title, template.HTML(page), fileName, fileName, rndrConf.SiteName, time.Now().Year(), rndrConf.SidebarLinks}
	// You pass base just by name, for some reason
	full, err := RenderTemplates("base.tmpl", &data, templates[:])
	if err != nil {
		return err
	}
	if _, found := strings.CutSuffix(saveTo, ".html"); !found {
		return saveToFile(full, fmt.Sprintf("%v.html", saveTo))
	}
	return saveToFile(full, saveTo)
}
