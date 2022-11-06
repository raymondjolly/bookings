package main

import (
	"bookings/internal/config"
	"bookings/internal/driver"
	"bookings/internal/handlers"
	"bookings/internal/helpers"
	"bookings/internal/models"
	"bookings/internal/render"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/fatih/color"
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
	db, err := run()
	errFatal(err)
	defer db.SQL.Close()

	defer close(app.MailChan)
	color.Cyan("Starting mail listener...")
	listenForMail()

	color.Green("Application has started on port %s", port)
	srv := &http.Server{Addr: port, Handler: routes(&app)}
	err = srv.ListenAndServe()
	errFatal(err)
}

func run() (*driver.DB, error) {
	//what am I going to put in the session?
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	//read flags
	inProduction := flag.Bool("production", true, "Application is in production")
	useCache := flag.Bool("cache", true, "Use template cache")
	dbHost := flag.String("dbhost", "localhost", "Database host")
	dbName := flag.String("dbname", "", "Database name")
	dbUser := flag.String("dbuser", "", "Database user")
	dbPass := flag.String("dbpass", "", "Database password")
	dbPort := flag.String("dbport", "5432", "Database port")
	dbSSL := flag.String("dbssl", "disable", "Database ssl settings (disable, prefer, required)")

	flag.Parse()

	if *dbName == "" || *dbUser == "" {
		fmt.Println("Missing required flags")
		os.Exit(1)
	}

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	app.InProduction = *inProduction
	app.UseCache = *useCache

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

	//connect to database
	color.Cyan("Connecting to database...")
	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s ",
		*dbHost, *dbPort, *dbName, *dbUser, *dbPass, *dbSSL)
	db, err := driver.ConnectSQL(connectionString)
	if err != nil {
		log.Fatalln("Cannot connect to the database. Dying.")
	}

	color.Green("Connected to database")
	templateCache, err := render.CreateTemplateCache()
	returnError(err)

	app.TemplateCache = templateCache

	repo := handlers.NewRepository(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}

func returnError(err error) error {
	return err
}

func errFatal(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}
