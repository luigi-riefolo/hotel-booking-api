package booking

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hotelbooking/internal/hotel"
)

// emailDelay simulates the time it takes to send the confirmation email
const emailDelay = 2 * time.Second

// Booking holds one room for every night of a stay
type Booking struct {
	CheckIn    time.Time  `gorm:"type:date;not null" json:"check_in"`
	CheckOut   time.Time  `gorm:"type:date;not null" json:"check_out"`
	CreatedAt  time.Time  `json:"created_at"`
	GuestEmail string     `gorm:"not null" json:"guest_email"`
	GuestName  string     `gorm:"not null" json:"guest_name"`
	Guests     int        `gorm:"not null" json:"guests"`
	ID         uint       `json:"-"`
	Reference  string     `gorm:"not null;uniqueIndex" json:"reference"`
	Room       hotel.Room `json:"room"`
	RoomID     uint       `gorm:"not null;index" json:"room_id"`
}

// Request is what a guest provides to book a room
type Request struct {
	RoomID     uint
	Guests     int
	GuestName  string
	GuestEmail string
	Stay       hotel.Stay
}

// Validate checks the parts of a request that need no database.
func (r *Request) Validate(now time.Time) error {

	if strings.TrimSpace(r.GuestName) == "" {
		return fmt.Errorf("guest_name is required")
	}

	if !strings.Contains(r.GuestEmail, "@") {
		return fmt.Errorf("guest_email '%s' is not an email address", r.GuestEmail)
	}

	if r.Guests < 1 {
		return fmt.Errorf("guests must be at least 1")
	}

	today := now.UTC().Truncate(24 * time.Hour)
	if r.Stay.CheckIn.Before(today) {
		return fmt.Errorf("check_in is in the past")
	}

	return nil
}

// Service books rooms and looks bookings up, all business logic goes here
type Service struct {
	DB *gorm.DB

	emails sync.WaitGroup
}

// Book reserves the room for the whole stay, or returns nil if it is already
// booked for any of those nights. The room row is locked for the length of
// the transaction,  so two requests for the same room are handled one after
// the other and cannot both pass the overlap check
func (s *Service) Book(ctx context.Context, rm hotel.Room, req Request) (*Booking, error) {

	b := &Booking{
		CheckIn:    req.Stay.CheckIn,
		CheckOut:   req.Stay.CheckOut,
		GuestEmail: strings.TrimSpace(req.GuestEmail),
		GuestName:  strings.TrimSpace(req.GuestName),
		Guests:     req.Guests,
		Reference:  newReference(),
		RoomID:     rm.ID,
	}

	taken := false

	// use transactions, so that in case of failures we can roll back
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&hotel.Room{}, rm.ID).Error
		if err != nil {
			return fmt.Errorf("could not lock room %d: %w", rm.ID, err)
		}

		clashes := int64(0)

		err = tx.Model(&Booking{}).
			Where("room_id = ? AND check_in < ? AND check_out > ?", rm.ID, b.CheckOut, b.CheckIn).
			Count(&clashes).Error
		if err != nil {
			return fmt.Errorf("could not check room %d for clashes: %w", rm.ID, err)
		}

		if clashes > 0 {
			taken = true
			return nil
		}

		return tx.Omit("Room").Create(b).Error
	})
	if err != nil {
		return nil, err
	}

	// nothing to do, sorry :(
	if taken {
		return nil, nil
	}

	b.Room = rm

	slog.Info("room booked", "ref", b.Reference, "room", b.RoomID)

	s.emails.Go(func() {
		sendConfirmation(b.Reference, b.GuestEmail)
	})

	return b, nil
}

// FindByReference returns the booking with the
// given reference, or nil if there is none
func (s *Service) FindByReference(ctx context.Context, reference string) (*Booking, error) {

	b := Booking{}

	err := s.DB.WithContext(ctx).
		Preload("Room").
		Where("reference = ?", strings.ToUpper(reference)).
		First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("could not load booking '%s': %w", reference, err)
	}

	return &b, nil
}

// Wait blocks until every pending confirmation email has been sent
func (s *Service) Wait() {

	s.emails.Wait()
}

// sendConfirmation pretends to email the guest (this is only a stub for now)
func sendConfirmation(reference, email string) {

	time.Sleep(emailDelay)

	slog.Info("confirmation email sent", "ref", reference, "to", email)
}

// newReference returns a random reference such as BK-3F9A01C2B4E57D10.
func newReference() string {

	b := make([]byte, 8)
	rand.Read(b)

	return fmt.Sprintf("BK-%X", b)
}
