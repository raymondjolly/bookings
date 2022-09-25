package repository

import (
	"bookings/internal/models"
	"time"
)

type DatabaseRepo interface {
	AllUsers() bool
	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(r models.RoomRestriction) error
	SearchAvailabilityByDatesByRoomID(startDate, endDate time.Time, roomId int) (bool, error)
	SearchAvailabilityForAllRooms(startDate, endDate time.Time) ([]models.Room, error)
	GetRoomById(id int) (models.Room, error)
	ModifyUser(u models.User) error
	Authenticate(email, testPassword string) (int, string, error)
	AllReservations() ([]models.Reservation, error)
	AllNewReservations() ([]models.Reservation, error)
}
