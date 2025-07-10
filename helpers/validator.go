package helpers

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
	"hris/config"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

type Validation struct {
	conn *sql.DB
}

func NewValidation() *Validation {
	conn, err := config.DBConnection()
	if err != nil {
		log.Println("Error validation :", err)
	}

	return &Validation{
		conn: conn,
	}
}

func (v *Validation) Init() (*validator.Validate, ut.Translator) {
	translator := en.New()
	uni := ut.New(translator, translator)
	trans, _ := uni.GetTranslator("en")
	validate := validator.New()

	// Register default translation
	en_translations.RegisterDefaultTranslations(validate, trans)

	// Set label field
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		labelName := field.Tag.Get("label")
		if labelName == "" {
			return field.Name
		}
		return labelName
	})

	// Custom translate "required"
	validate.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0} tidak boleh kosong", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())
		return t
	})

	// Custom translate "email"
	validate.RegisterTranslation("email", trans, func(ut ut.Translator) error {
		return ut.Add("email", "{0} harus berupa email yang valid", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("email", fe.Field())
		return t
	})

	// Custom translate "gte"
	validate.RegisterTranslation("gte", trans, func(ut ut.Translator) error {
		return ut.Add("gte", "{0} minimal harus {1} karakter", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("gte", fe.Field(), fe.Param())
		return t
	})

	// Custom translate "eqfield" (cocokkan field,)
	validate.RegisterTranslation("eqfield", trans, func(ut ut.Translator) error {
		return ut.Add("eqfield", "{0} harus sama dengan {1}", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("eqfield", fe.Field(), fe.Param())
		return t
	})

	// Register validation "isunique"
	validate.RegisterValidation("isunique", func(fl validator.FieldLevel) bool {
		params := fl.Param()
		splitParam := strings.Split(params, "-")

		tableName := splitParam[0]
		fieldName := splitParam[1]
		fieldValue := fl.Field().String()

		return v.checkIsUnique(tableName, fieldName, fieldValue)
	})

	// Custom translate "isunique"
	validate.RegisterTranslation("isunique", trans, func(ut ut.Translator) error {
		return ut.Add("isunique", "{0} sudah digunakan", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("isunique", fe.Field())
		return t
	})

	return validate, trans
}

func (v *Validation) Struct(s interface{}) interface{} {
	validate, trans := v.Init()
	var vErrors = make(map[string]interface{})

	err := validate.Struct(s)
	if err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			vErrors[e.StructField()] = e.Translate(trans)
		}
	}

	if len(vErrors) > 0 {
		return vErrors
	}

	return nil
}

func (v *Validation) checkIsUnique(tableName, fieldName, fieldValue string) bool {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", fieldName, tableName, fieldName)
	row := v.conn.QueryRow(query, fieldValue)

	var result string
	err := row.Scan(&result)

	// Jika tidak ditemukan (err == sql.ErrNoRows), berarti unik → return true
	if err == sql.ErrNoRows {
		return true
	}

	// Jika ditemukan, berarti tidak unik → return false
	return false
}