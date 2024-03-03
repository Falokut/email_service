package models

import "time"

type Place struct {
	Row  int32 `bson:"row" json:"row"`
	Seat int32 `bson:"seat" json:"seat"`
}

type Ticket struct {
	Id     string          `json:"id"`
	Place  Place           `json:"place"`
	Price  uint32          `json:"price"`
}

type Order struct {
	Id          string    `json:"id"`
	Tickets     []Ticket  `json:"tickets"`
	ScreeningId int64     `json:"screening_id"`
	Date        time.Time `json:"order_date"`
}
