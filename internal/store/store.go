package store

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"hotelbooking/internal/booking"
	"hotelbooking/internal/hotel"
)

func Open(dsn string) (*gorm.DB, error) {

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if err != nil {
		return nil, fmt.Errorf("could not connect to the database: %w", err)
	}

	err = db.AutoMigrate(&hotel.Hotel{}, &hotel.Room{}, &booking.Booking{})
	if err != nil {
		return nil, fmt.Errorf("could not migrate the schema: %w", err)
	}

	return db, nil
}

// every hotel gets the same six rooms, two of each type
var rooms = []hotel.Room{
	{Number: "101", RoomType: hotel.Single, Capacity: 1},
	{Number: "102", RoomType: hotel.Single, Capacity: 1},
	{Number: "201", RoomType: hotel.Double, Capacity: 2},
	{Number: "202", RoomType: hotel.Double, Capacity: 2},
	{Number: "301", RoomType: hotel.Deluxe, Capacity: 4},
	{Number: "302", RoomType: hotel.Deluxe, Capacity: 4},
}

// Seed resets the database and fills it with three hotels
// TODO: maybe use sidecar or init script?
func Seed(ctx context.Context, db *gorm.DB) ([]hotel.Hotel, error) {

	err := Reset(ctx, db)
	if err != nil {
		return nil, err
	}

	hotels := []hotel.Hotel{}

	for _, name := range []string{"The Overlook", "Grand Budapest", "Bates Motel"} {
		hotels = append(hotels, hotel.Hotel{Name: name, Rooms: append([]hotel.Room{}, rooms...)})
	}

	err = db.WithContext(ctx).Create(&hotels).Error
	if err != nil {
		return nil, fmt.Errorf("could not seed the hotels: %w", err)
	}

	slog.Info("database seeded", "hotels", len(hotels))

	return hotels, nil
}

// Reset removes all data and restarts the ids
func Reset(ctx context.Context, db *gorm.DB) error {

	err := db.WithContext(ctx).Exec("TRUNCATE bookings, rooms, hotels RESTART IDENTITY").Error
	if err != nil {
		return fmt.Errorf("could not reset the database: %w", err)
	}

	slog.Info("database reset")

	return nil
}
