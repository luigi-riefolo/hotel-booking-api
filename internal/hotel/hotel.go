package hotel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// RoomType is one of the three room types every hotel offers
type RoomType string

const (
	Single RoomType = "single"
	Double RoomType = "double"
	Deluxe RoomType = "deluxe"
)

// Hotel has  six rooms
type Hotel struct {
	ID    uint   `json:"id"`
	Name  string `gorm:"not null;uniqueIndex" json:"name"`
	Rooms []Room `json:"rooms"`
}

type Room struct {
	ID       uint     `json:"id"`
	HotelID  uint     `gorm:"not null;index" json:"hotel_id"`
	Number   string   `gorm:"not null" json:"number"`
	RoomType RoomType `gorm:"not null" json:"type"`
	Capacity int      `gorm:"not null" json:"capacity"`
}

type Service struct {
	DB *gorm.DB
}

// SearchByName returns the hotels whose name contains the given text,
// ignoring case, together with their rooms.
func (s *Service) SearchByName(ctx context.Context, name string) ([]Hotel, error) {

	hotels := []Hotel{}

	err := s.DB.WithContext(ctx).
		Preload("Rooms").
		Where("name ILIKE ?", "%"+name+"%").
		Order("name").
		Find(&hotels).Error
	if err != nil {
		return nil, fmt.Errorf("could not search hotels by name '%s': %w", name, err)
	}

	return hotels, nil
}

// AvailableRooms returns the rooms of a hotel that can sleep the party and
// are free for every night of the stay,  so guests never have to change room.
func (s *Service) AvailableRooms(ctx context.Context, hotelID uint, st Stay, guests int) ([]Room, error) {

	rooms := []Room{}

	err := s.DB.WithContext(ctx).
		Where("hotel_id = ? AND capacity >= ?", hotelID, guests).
		Where(`NOT EXISTS (
			SELECT 1 FROM bookings b
			WHERE b.room_id = rooms.id AND b.check_in < ? AND b.check_out > ?
		)`, st.CheckOut, st.CheckIn).
		Order("number").
		Find(&rooms).Error
	if err != nil {
		return nil, fmt.Errorf("could not find available rooms in hotel %d: %w", hotelID, err)
	}

	return rooms, nil
}

// FindRoom returns the room with the given id, or nil if there is none.
func (s *Service) FindRoom(ctx context.Context, id uint) (*Room, error) {

	rm := Room{}

	err := s.DB.WithContext(ctx).First(&rm, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("could not load room %d: %w", id, err)
	}

	return &rm, nil
}

// Stay is the range of nights a guest sleeps in a room
type Stay struct {
	CheckIn  time.Time
	CheckOut time.Time
}

// ParseStay reads a pair of YYYY-MM-DD dates into a stay of at least one night
func ParseStay(checkIn, checkOut string) (Stay, error) {

	in, err := time.Parse(time.DateOnly, checkIn)
	if err != nil {
		return Stay{}, fmt.Errorf("could not parse check_in '%s', expected YYYY-MM-DD", checkIn)
	}

	out, err := time.Parse(time.DateOnly, checkOut)
	if err != nil {
		return Stay{}, fmt.Errorf("could not parse check_out '%s', expected YYYY-MM-DD", checkOut)
	}

	if !out.After(in) {
		return Stay{}, fmt.Errorf("check_out '%s' must be after check_in '%s'", checkOut, checkIn)
	}

	return Stay{CheckIn: in, CheckOut: out}, nil
}
