package entities

type Shift struct {
	Id        int64
	Name      string `validate:"required" label:"Nama Shift"`
	StartTime string `validate:"required" label:"Jam Masuk"`
	EndTime   string `validate:"required" label:"Jam Pulang"`
}