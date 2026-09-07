package domain

import "errors"

var (
	ErrEmailAlreadyRegistered   = errors.New("el correo ya esta registrado")
	ErrEmailPendingVerification = errors.New("la cuenta existe pero no ha sido verificada")
)

var (
	ErrBusinessNameRequired     = errors.New("el nombre del negocio es obligatorio")
	ErrGoogleSignupTokenInvalid = errors.New("la solicitud de registro con Google expiró, intenta de nuevo")
)
