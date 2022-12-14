package repository

import (
	"bookings/internal/models"
	"time"
)

type DatabaseRepo interface {
	AllUsers() bool
	AllRooms() ([]models.Room, error)

	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(r models.RoomRestriction) error
	SearchAvailabilityByDatesByRoomID(startDate, endDate time.Time, roomId int) (bool, error)
	SearchAvailabilityForAllRooms(startDate, endDate time.Time) ([]models.Room, error)

	GetRoomById(id int) (models.Room, error)
	UpdateUser(u models.User) error

	Authenticate(email, testPassword string) (int, string, error)

	AllReservations() ([]models.Reservation, error)
	AllNewReservations() ([]models.Reservation, error)

	GetReservationById(id int) (models.Reservation, error)
	UpdateReservation(u models.Reservation) error
	UpdateProcessedForReservation(id, processed int) error
	DeleteReservation(id int) error

	InsertBlockForRoom(id int, startDate time.Time) error
	DeleteBlockById(id int) error

	GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error)
}
