package wiki

import (
	"encoding/json"
	"github.com/gosimple/slug"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Load(path, projectName string, sidebars []*Sidebar, projectNameToWikiIndex map[string]WikisByName, projectNameToProjectSlug map[string]string) ([]*Wiki, error) {
	var result []*Wiki

	fileList := make([]string, 0)
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return err
	})

	if err != nil {
		return nil, err
	}

	for _, file := range fileList {
		if strings.HasSuffix(file, ".md") {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				return nil, err
			}

			page := new(Wiki)
			page.Name = filepath.Base(strings.ReplaceAll(file, ".md", ""))
			page.Slug = slug.Make(page.Name)
			page.Body = parseBody(string(data))

			metaFile := strings.ReplaceAll(file, ".md", ".json")
			if _, err := os.Stat(metaFile); err == nil {
				data, err := ioutil.ReadFile(metaFile)
				if err != nil {
					return nil, err
				}

				err = json.Unmarshal(data, &page.Meta)
				if err != nil {
					return nil, err
				}
			}

			result = append(result, page)

			projectNameToWikiIndex[projectName][page.Name] = page
		}
	}

	for _, sidebar := range sidebars {
		data, err := ioutil.ReadFile(sidebar.File)
		if err != nil {
			return nil, err
		}

		sidebar.Body = parseBody(string(data))
	}

	for _, page := range result {
		result, err := parseReferenceLinks(page.Body, page.Name, projectName, referencesAndVariables, projectNameToWikiIndex, projectNameToProjectSlug)
		if err != nil {
			return nil, err
		}

		page.Body = result
	}

	for _, page := range result {
		result, err := parseReferenceLinks(page.Body, page.Name, projectName, includes, projectNameToWikiIndex, projectNameToProjectSlug)
		if err != nil {
			return nil, err
		}

		page.Body = result
	}

	for _, sidebar := range sidebars {
		result, err := parseReferenceLinks(sidebar.Body, sidebar.Name, projectName, referencesAndVariables, projectNameToWikiIndex, projectNameToProjectSlug)
		if err != nil {
			return nil, err
		}

		sidebar.BodyHtml = template.HTML(result)
	}

	return result, nil
}

func (w *Wiki) PostLoad(projectName string, projectNameToProjectSlug map[string]string, projectNameToWikiIndex map[string]WikisByName) error {
	result, err := parseReferenceLinks(w.Body, w.Name, projectName, crossProjectReferences, projectNameToWikiIndex, projectNameToProjectSlug)
	if err != nil {
		return err
	}

	w.Body = result

	return nil
}
