package handlers

import (
	"bookings/internal/config"
	"bookings/internal/driver"
	"bookings/internal/forms"
	"bookings/internal/helpers"
	"bookings/internal/models"
	"bookings/internal/render"
	"bookings/internal/repository"
	"bookings/internal/repository/dbrepo"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
	"time"
)

var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepository creates a new Repository
func NewRepository(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		a,
		dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewTestRepo creates a new Testing Repository
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		a,
		dbrepo.NewTestingRepo(a),
	}
}

func NewHandlers(r *Repository) {
	Repo = r
}

// Home is a 'home' route function
func (rep *Repository) Home(w http.ResponseWriter, r *http.Request) {
	_ = render.Template(w, r, "home.page.tmpl", &models.TemplateData{})

}

// About is the 'about' page handler
func (rep *Repository) About(w http.ResponseWriter, r *http.Request) {
	_ = render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

func (rep *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	_ = render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

func (rep *Repository) ColonelsSuite(w http.ResponseWriter, r *http.Request) {
	_ = render.Template(w, r, "colonels-suite.page.tmpl", &models.TemplateData{})
}

func (rep *Repository) GeneralsQuarters(w http.ResponseWriter, r *http.Request) {
	_ = render.Template(w, r, "generals-quarters.page.tmpl", &models.TemplateData{})
}

// PostReservation handles the posting of a reservation form
func (rep *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "cannot parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")
	layout := "2006-01-02"

	startDate, err := time.Parse(layout, sd)
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "cannot parse start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
	endDate, err := time.Parse(layout, ed)
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "cannot parse end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "invalid data")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	room, err := rep.DB.GetRoomById(roomID)
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "cannot find room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Phone:     r.Form.Get("phone"),
		Email:     r.Form.Get("email"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    roomID,
		Room:      room,
	}

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3)
	form.MinLength("last_name", 2)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		http.Error(w, "my own error message", http.StatusSeeOther)
		_ = render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationID, err := rep.DB.InsertReservation(reservation)
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "cannot insert reservation into the database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}
	err = rep.DB.InsertRoomRestriction(restriction)
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "cannot insert restriction into the database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	//send notifications - first to guest
	htmlMessage := fmt.Sprintf(`
	<strong>Reservation Confirmation</strong>
	Dear %sa:<br>
	This is to confirm your reservation fromn %s to %s
`, reservation.FirstName, reservation.StartDate.Format(layout), reservation.EndDate.Format(layout))

	msg := models.MailData{
		To:       "me@here.com",
		From:     "me@here.com",
		Subject:  "Reservation Notification",
		Content:  htmlMessage,
		Template: "basic.html",
	}

	rep.App.MailChan <- msg
	//send notifications to property manager
	//send notifications - first to guest
	htmlMessage = fmt.Sprintf(`
	<strong>Reservation Confirmation</strong>
	Dear %sa:<br>
	This is to confirm a reservation from %s from %s to %s.
`, reservation.FirstName, reservation.Room.RoomName, reservation.StartDate.Format(layout), reservation.EndDate.Format(layout))

	rep.App.MailChan <- msg

	rep.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

func (rep *Repository) MakeReservation(w http.ResponseWriter, r *http.Request) {
	res, ok := rep.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		rep.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := rep.DB.GetRoomById(res.RoomID)
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "can't find room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	res.Room.RoomName = room.RoomName

	rep.App.Session.Put(r.Context(), "reservation", res)

	layout := "2006-01-02"
	sd := res.StartDate.Format(layout)
	ed := res.EndDate.Format(layout)

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	_ = render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

func (rep *Repository) SearchAvailability(w http.ResponseWriter, r *http.Request) {
	_ = render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

func (rep *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	checkServerError(w, err)
	endDate, err := time.Parse(layout, end)
	checkServerError(w, err)

	rooms, err := rep.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	checkServerError(w, err)

	if len(rooms) == 0 {
		rep.App.Session.Put(r.Context(), "error", "No rooms available")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
	}
	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}
	rep.App.Session.Put(r.Context(), "reservation", res)

	_ = render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomId    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON handles the request for availability and sends JSON response
func (rep *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		//can't parse form, so return appropriate json
		resp := jsonResponse{
			OK:      false,
			Message: "Internal server error",
		}

		out, _ := json.MarshalIndent(resp, "", "    ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")
	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	roomId, err := strconv.Atoi(r.Form.Get("room_id"))
	checkParseError(err)
	available, err := rep.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomId)
	if err != nil {
		//can't parse form, so return appropriate json
		resp := jsonResponse{
			OK:      false,
			Message: "Error connecting to database",
		}

		out, _ := json.MarshalIndent(resp, "", "    ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}
	resp := jsonResponse{
		OK:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomId:    strconv.Itoa(roomId),
	}

	//error check removed since data is handled with json
	out, _ := json.MarshalIndent(resp, "", "     ")

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// ReservationSummary displays the reservation summary page
func (rep *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := rep.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		rep.App.Session.Put(r.Context(), "error", "Cannot get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	rep.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation
	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	_ = render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
	fmt.Println(reservation.Room)
}

// ChooseRoom displays available rooms
func (rep *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomId, err := strconv.Atoi(chi.URLParam(r, "id"))
	checkServerError(w, err)

	res, ok := rep.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	checkErrorOk(w, ok, "cannot get session data")

	res.RoomID = roomId
	rep.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// BookRoom takes URL parameters, builds a session variable and takes user to make res screen
func (rep *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	//id, s, e
	roomId, _ := strconv.Atoi(r.URL.Query().Get("id"))
	s := r.URL.Query().Get("s")
	e := r.URL.Query().Get("e")

	format := "2006-01-02"
	var res models.Reservation
	room, err := rep.DB.GetRoomById(roomId)
	if err != nil {
		rep.App.Session.Put(r.Context(), "error", "cannot get room from db")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	res.Room.RoomName = room.RoomName
	res.RoomID = roomId
	res.StartDate, _ = time.Parse(format, s)
	res.EndDate, _ = time.Parse(format, e)

	rep.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)

}

func (rep *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostShowLogin handles logging the user in
func (rep *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	// prevents session fixation attack. Always use this for login and logout
	_ = rep.App.Session.RenewToken(r.Context())
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := rep.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)
		rep.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	rep.App.Session.Put(r.Context(), "user_id", id)
	rep.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout logs a user out
func (rep *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = rep.App.Session.Destroy(r.Context())
	_ = rep.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (rep *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

func (rep *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{})
}

func (rep *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{})
}

func (rep *Repository) AdminCalendarReservations(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{})
}

func checkServerError(w http.ResponseWriter, err error) {
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
}

func checkErrorOk(w http.ResponseWriter, ok bool, errDesc string) {
	if !ok {
		helpers.ServerError(w, errors.New(errDesc))
	}
}

func checkParseError(err error) {
	if err != nil {
		helpers.ParseError(err)
	}
}
