package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

type templateData struct {
	Data map[string]any
	API  string
}

//go:embed views
var templateFS embed.FS

func (app *application) render(w http.ResponseWriter, page string, td *templateData) error {
	templateFile := fmt.Sprintf("views/%s.gohtml", page)

	t, err := app.parse(page, templateFile)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	td.API = app.ApiAddr

	err = t.Execute(w, td)

	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	return nil
}

func (app *application) parse(page, templateFile string) (*template.Template, error) {
	funcs := template.FuncMap{
		"parseResult": parseResult,
	}

	t, err := template.New(fmt.Sprintf("%s.gohtml", page)).
		Funcs(funcs).
		ParseFS(templateFS, "views/base.gohtml", templateFile)

	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}

	return t, nil
}

func parseResult(resultStr string) map[string]any {
	var result map[string]any
	if err := json.Unmarshal([]byte(resultStr), &result); err != nil {
		return map[string]any{"error": "failed to parse result"}
	}
	return result
}
