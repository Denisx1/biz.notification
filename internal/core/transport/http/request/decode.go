package core_http_request

import (
	"encoding/json"
	"fmt"
	"net/http"
	core_errors "notification/internal/core/errors"

	"gopkg.in/go-playground/validator.v9"
)

var requestValidater = validator.New()

type validateble interface {
	Validate() error
}

func DecodeAndValidateRequest(r *http.Request, dest any) error {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		return fmt.Errorf("decode json: %v: %w", err, core_errors.ErrInvalidArgument)

	}
	var (
		err error
	)
	v, ok := dest.(validateble) // type assertion позволяет провертиь подходит ли стуктура пол интерфейс
	if ok {                     // еслди е сть кастомные правила валидации то тогда применяются елс и не т то идем дальши и валиджируем постраринке
		err = v.Validate()
	} else {
		err = requestValidater.Struct(dest)
	}
	if err != nil {
		return fmt.Errorf(
			"request validation: %v: %w",
			err,
			core_errors.ErrInvalidArgument)
	}

	return nil
}
