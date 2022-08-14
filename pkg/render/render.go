package render

import (
	"bookings/pkg/config"
	"bookings/pkg/models"
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var app *config.AppConfig

//NewTemplates sets the config for the template package
func NewTemplates(a *config.AppConfig) {
	app = a
}

func AddDefaultDate(td *models.TemplateData) *models.TemplateData {
	return td
}

//RenderTemplate renders a template
func RenderTemplate(w http.ResponseWriter, tmpl string, td *models.TemplateData) {
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
		log.Fatalln("Could not get template from templateCache")
	}

	buf := new(bytes.Buffer)
	td = AddDefaultDate(td)
	err := t.Execute(buf, td)
	errTemplate(err)

	//render the template
	_, err = buf.WriteTo(w)
	errTemplate(err)
}

//CreateTemplateCache creates a template cache
func CreateTemplateCache() (map[string]*template.Template, error) {
	templateCache := map[string]*template.Template{}

	//get all of the files named *.page.tmpl
	pages, err := filepath.Glob("./templates/*.page.tmpl")
	if err != nil {
		return templateCache, err
	}

	//range through all files ending with *.page.tmpl
	for _, page := range pages {
		name := filepath.Base(page)
		//parse file
		templateSet, err := template.New(name).ParseFiles(page)
		if err != nil {
			return templateCache, err
		}
		//now look for layouts in the directory
		matches, err := filepath.Glob("./templates/*layout.tmpl")
		if err != nil {
			return templateCache, err
		}
		if len(matches) > 0 {
			templateSet, err = templateSet.ParseGlob("./templates/*layout.tmpl")
			if err != nil {
				return templateCache, err
			}
			templateCache[name] = templateSet
		}
	}

	return templateCache, nil
}

func errTemplate(e error) {
	if e != nil {
		log.Println("error parsing template", e)
	}
}

func errFatal(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}
