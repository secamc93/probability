package googleoauth

import "errors"

var (
	ErrNotConfigured  = errors.New("el inicio de sesión con Google no está configurado")
	ErrExchangeFailed = errors.New("no se pudo validar la cuenta de Google")
)
