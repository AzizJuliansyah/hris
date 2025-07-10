package entities

type Office struct {
	Id        int64
	Name      string  `validate:"required" label:"Nama Kantor"`
	Address   string  `validate:"required" label:"Alamat Kantor"`
	Latitude  float64 `validate:"required"`
	Longitude float64 `validate:"required"`
	Radius    int64   `validate:"required"`
}