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
	"github.com/go-chi/chi/v5"
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

func NewHandlers(r *Repository) {
	Repo = r
}

// Home is a 'home' route function
func (rep *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})

}

// About is the 'about' page handler
func (rep *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

func (rep *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

func (rep *Repository) ColonelsSuite(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "colonels-suite.page.tmpl", &models.TemplateData{})
}

func (rep *Repository) GeneralsQuarters(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals-quarters.page.tmpl", &models.TemplateData{})
}

// PostReservation handles the posting of a reservation form
func (rep *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	//01/02 03:04:05PM '06 -0700
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, sd)
	checkServerError(w, err)
	endDate, err := time.Parse(layout, ed)
	checkServerError(w, err)
	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	checkServerError(w, err)

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
		RoomID:    roomID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	form := forms.New(r.PostForm)
	//form.Has("first_name", r)
	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 3)
	form.MinLength("last_name", 2)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationID, err := rep.DB.InsertReservation(reservation)
	checkServerError(w, err)

	restriction := models.RoomRestriction{
		StartDate:     startDate,
		EndDate:       endDate,
		RoomID:        roomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}
	err = rep.DB.InsertRoomRestriction(restriction)
	checkServerError(w, err)

	rep.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

func (rep *Repository) MakeReservation(w http.ResponseWriter, r *http.Request) {
	res, ok := rep.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	checkErrorOk(w, ok, "cannot get reservation from session")
	//if !ok {
	//	helpers.ServerError(w, errors.New("cannot get reservation from session"))
	//}
	room, err := rep.DB.GetRoomById(res.RoomID)
	checkServerError(w, err)
	res.Room.RoomName = room.RoomName

	layout := "2006-01-02"
	sd := res.StartDate.Format(layout)
	ed := res.EndDate.Format(layout)

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

func (rep *Repository) SearchAvailability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
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

	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func (rep *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse{OK: true, Message: "available"}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (rep *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := rep.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		rep.App.ErrorLog.Println("cannot get error from session")
		rep.App.Session.Put(r.Context(), "warning", "Cannot get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	rep.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

func (rep *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomId, err := strconv.Atoi(chi.URLParam(r, "id"))
	checkServerError(w, err)

	res, ok := rep.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	checkErrorOk(w, ok, "cannot get session data")

	res.RoomID = roomId
	rep.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
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
