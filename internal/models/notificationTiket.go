package models

type TicketNotification struct {
	Id        string
	IdBarCode string
	Row       int32
	Seat      int32
	Price     string
}
