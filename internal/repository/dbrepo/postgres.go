package dbrepo

import (
	"bookings/internal/models"
	"context"
	"time"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts a reservation into the database
func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var newID int

	stmt := `insert into reservations (first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at) 
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	err := m.DB.QueryRowContext(ctx, stmt, res.FirstName, res.LastName, res.Email, res.Phone,
		res.StartDate, res.EndDate, res.RoomID, time.Now(), time.Now()).Scan(&newID)
	if err != nil {
		return 0, err
	}
	return newID, nil
}

// InsertRoomRestriction inserts a room restriction into the database
func (m *postgresDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into room_restrictions(start_date, end_date, room_id, created_at, updated_at, reservation_id, restriction_id) 
			values ($1, $2, $3, $4, $5, $6, $7)`
	_, err := m.DB.ExecContext(ctx, stmt, res.StartDate, res.EndDate, res.RoomID,
		time.Now(), time.Now(), res.ReservationID, res.RestrictionID)
	if err != nil {
		return err
	}
	return nil
}

// SearchAvailabilityByDates returns true if an availablity exists and false if no availability exits
func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(startDate, endDate time.Time, roomId int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var numOfRows int

	query := `select count(id) from room_restrictions where $1 < end_date and $2 > start_date and room_id = $3`
	row := m.DB.QueryRowContext(ctx, query, startDate, endDate, roomId)
	err := row.Scan(&numOfRows)
	if err != nil {
		return false, err
	}
	if numOfRows == 0 {
		return true, nil
	}
	return false, nil
}

// SearchAvailabilityForAllRooms returns a slice of available rooms, if any for a given date range
func (m *postgresDBRepo) SearchAvailabilityForAllRooms(startDate, endDate time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var rooms []models.Room

	query := `
		select r.id, r.room_name
		from rooms r
		where r.id not in (select rr.room_id from room_restrictions rr where $1 <rr.end_date and $2 >rr.start_date)`

	rows, err := m.DB.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return rooms, err
	}
	for rows.Next() {
		var room models.Room
		err := rows.Scan(
			&room.ID,
			&room.RoomName,
		)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}
	return rooms, nil
}

// GetRoomById gets a room by id
func (m *postgresDBRepo) GetRoomById(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var room models.Room

	query := `select id, room_name, created_at, updated_at from rooms where id =$1`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&room.ID,
		&room.RoomName,
		&room.CreatedAt,
		&room.UpdatedAt)
	if err != nil {
		return room, err
	} else {
		return room, nil
	}
}
