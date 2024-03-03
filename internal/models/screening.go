package models

type Screening struct {
	// formated like hh:mm
	StartTime string
	// formated like  dd.mm
	StartDate string

	MovieName      string
	MoviePosterUrl string
	Cinema         Cinema
	HallName       string
}
