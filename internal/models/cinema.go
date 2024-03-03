package models

type Coordinates struct {
	Long float64
	Lat  float64
}

type Cinema struct {
	Address     string
	Name        string
	Coordinates Coordinates
}
