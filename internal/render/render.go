package render

import (
	"bookings/internal/config"
	"bookings/internal/models"
	"bytes"
	"errors"
	"fmt"
	"github.com/justinas/nosurf"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

var functions = template.FuncMap{
	"humanDate":  HumanDate,
	"formatDate": FormatDate,
	"iterate":    Iterate,
	"add":        Add,
}

var app *config.AppConfig
var pathToTemplates = "./templates"

// NewRenderer sets the config for the template package
func NewRenderer(a *config.AppConfig) {
	app = a
}

func Add(a, b int) int {
	return a + b
}

// Iterate returns a slice of ins, starting at 1, going to count
func Iterate(count int) []int {
	var i int
	var items []int
	for i = 0; i < count; i++ {
		items = append(items, i)
	}
	return items
}

// HumanDate returns time in YYY-MM-DD format
func HumanDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func FormatDate(t time.Time, f string) string {
	return t.Format(f)
}
func AddDefaultDate(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.CSRFToken = nosurf.Token(r)
	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1
	}
	return td
}

// Template renders a template
func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {
	var tc = map[string]*template.Template{}
	if app.UseCache {
		//create a template cache
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	//get requested template from cache
	t, okay := tc[tmpl]
	if !okay {
		return errors.New("cannot get template from cache")
	}

	buf := new(bytes.Buffer)
	td = AddDefaultDate(td, r)
	err := t.Execute(buf, td)
	errTemplate(err)

	//render the template
	_, err = buf.WriteTo(w)
	errTemplate(err)
	return nil
}

// CreateTemplateCache creates a template cache
func CreateTemplateCache() (map[string]*template.Template, error) {
	templateCache := map[string]*template.Template{}

	//get all files named *.page.tmpl
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return templateCache, err
	}

	//range through all files ending with *.page.tmpl
	for _, page := range pages {
		name := filepath.Base(page)
		//parse file
		templateSet, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return templateCache, err
		}
		//now look for layouts in the directory
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return templateCache, err
		}
		if len(matches) > 0 {
			templateSet, err = templateSet.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
			if err != nil {
				return templateCache, err
			}
			templateCache[name] = templateSet
		}
	}

	return templateCache, nil
}

func errTemplate(e error) error {
	if e != nil {
		log.Println("error parsing template", e)
	}
	return e
}

func errFatal(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}
