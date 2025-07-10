package entities

type Auth struct{
	NIK		 string `validate:"required"`
	Password string `validate:"required"`
}