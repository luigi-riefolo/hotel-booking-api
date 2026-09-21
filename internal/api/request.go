package api

type BookingRequest struct {
	RoomID     uint   `json:"room_id"`
	Guests     int    `json:"guests"`
	GuestName  string `json:"guest_name"`
	GuestEmail string `json:"guest_email"`
	CheckIn    string `json:"check_in"`
	CheckOut   string `json:"check_out"`
}
