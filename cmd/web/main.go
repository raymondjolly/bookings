package main

import (
	"bookings/internal/config"
	"bookings/internal/handlers"
	"bookings/internal/helpers"
	"bookings/internal/models"
	"bookings/internal/render"
	"encoding/gob"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"log"
	"net/http"
	"os"
	"time"
)

const port = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

func main() {
	err := run()
	errFatal(err)

	fmt.Println("Application has started on port", port)
	srv := &http.Server{Addr: port, Handler: routes(&app)}
	err = srv.ListenAndServe()
	errFatal(err)
}

func run() error {
	//what am I going to put in the session?
	gob.Register(models.Reservation{})

	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ldate|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	app.Session = session

	templateCache, err := render.CreateTemplateCache()
	returnError(err)

	app.TemplateCache = templateCache
	app.UseCache = true

	repo := handlers.NewRepository(&app)
	handlers.NewHandlers(repo)
	render.NewTemplates(&app)
	helpers.NewHelpers(&app)

	return nil
}

func returnError(err error) error {
	return err
}

func errFatal(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}
