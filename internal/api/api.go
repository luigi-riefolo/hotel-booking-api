package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"hotelbooking/internal/booking"
	"hotelbooking/internal/hotel"
	"hotelbooking/internal/store"
)

// Server maps the hotel and booking services onto HTTP
type Server struct {
	Store           *store.Store
	Hotels          *hotel.Service
	BookingsService *booking.Service
}

func (s *Server) Routes() *echo.Echo {

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.RequestLogger())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	v1 := e.Group("/api/v1")

	v1.GET("/hotels", s.searchHotels)
	v1.GET("/hotels/:id/rooms", s.availableRooms)
	v1.POST("/bookings", s.book)
	v1.GET("/bookings/:reference", s.getBooking)

	// additional routs for seeding and rest
	v1.POST("/test/seed", s.seed)
	v1.POST("/test/reset", s.reset)

	return e
}

func (s *Server) searchHotels(c echo.Context) error {

	name := strings.TrimSpace(c.QueryParam("name"))
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}

	hotels, err := s.Hotels.SearchByName(c.Request().Context(), name)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, hotels)
}

func (s *Server) availableRooms(c echo.Context) error {

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "hotel id must be a number")
	}

	st, err := hotel.ParseStay(c.QueryParam("check_in"), c.QueryParam("check_out"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	guests, err := strconv.Atoi(c.QueryParam("guests"))
	if err != nil || guests < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "guests must be a number of at least 1")
	}

	rooms, err := s.Hotels.AvailableRooms(c.Request().Context(), uint(id), st, guests)
	if err != nil {
		// TODO: maybe return a bit more details about the type of error
		return err
	}

	return c.JSON(http.StatusOK, rooms)
}

func (s *Server) book(c echo.Context) error {

	req := BookingRequest{}

	err := c.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "body must be valid JSON")
	}

	st, err := hotel.ParseStay(req.CheckIn, req.CheckOut)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	br := booking.Request{
		RoomID:     req.RoomID,
		Guests:     req.Guests,
		GuestName:  req.GuestName,
		GuestEmail: req.GuestEmail,
		Stay:       st,
	}

	err = br.Validate(time.Now())
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()

	rm, err := s.Hotels.FindRoom(ctx, br.RoomID)
	if err != nil {
		return err
	}
	if rm == nil {
		return echo.NewHTTPError(http.StatusNotFound, "room not found")
	}

	if br.Guests > rm.Capacity {
		return echo.NewHTTPError(http.StatusUnprocessableEntity,
			fmt.Sprintf("room %d sleeps at most %d guests", rm.ID, rm.Capacity))
	}

	b, err := s.BookingsService.Book(ctx, *rm, br)
	if err != nil {
		return err
	}

	if b == nil {
		return echo.NewHTTPError(http.StatusConflict, "room is already booked for those nights")
	}

	return c.JSON(http.StatusCreated, b)
}

func (s *Server) getBooking(c echo.Context) error {

	b, err := s.BookingsService.FindByReference(c.Request().Context(), c.Param("reference"))
	if err != nil {
		return err
	}

	if b == nil {
		return echo.NewHTTPError(http.StatusNotFound, "booking not found")
	}

	return c.JSON(http.StatusOK, b)
}

func (s *Server) seed(c echo.Context) error {

	hotels, err := s.Store.Seed(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, hotels)
}

func (s *Server) reset(c echo.Context) error {

	err := s.Store.Reset(c.Request().Context())
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
