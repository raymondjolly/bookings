package models

import (
	"time"
)

// User is the user model
type User struct {
	ID          int       `json:"id"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	AccessLevel int       `json:"accessLevel"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Room is the room model
type Room struct {
	ID        int       `json:"ID"`
	RoomName  string    `json:"roomName"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Restriction struct {
	ID              int       `json:"ID"`
	RestrictionName string    `json:"roomName"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type Reservation struct {
	ID        int       `json:"ID"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	RoomID    int       `json:"roomID"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Room      Room
}

// RoomRestriction is the room restriction model
type RoomRestriction struct {
	ID            int
	StartDate     time.Time `json:"startDate"`
	EndDate       time.Time `json:"endDate"`
	RoomID        int
	ReservationID int
	RestrictionID int
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Room          Room
	Reservation   Reservation
	Restriction   Restriction
}

// MailData holds an email message
type MailData struct {
	To       string
	From     string
	Subject  string
	Content  string
	Template string
}
