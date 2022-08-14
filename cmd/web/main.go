package main

import (
	"bookings/pkg/config"
	"bookings/pkg/handlers"
	"bookings/pkg/render"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"log"
	"net/http"
	"time"
)

const port = ":8080"

var app config.AppConfig
var session *scs.SessionManager

func main() {

	app.InProduction = false
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session
	templateCache, err := render.CreateTemplateCache()
	errFatal(err)

	app.TemplateCache = templateCache
	app.UseCache = true

	repo := handlers.NewRepository(&app)
	handlers.NewHandlers(repo)
	render.NewTemplates(&app)

	fmt.Println("Application has started on port", port)
	srv := &http.Server{Addr: port, Handler: routes(&app)}
	err = srv.ListenAndServe()
	errFatal(err)
}

func errCheck(e error) {
	if e != nil {
		log.Panicln(e)
	}
}

func errFatal(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}
