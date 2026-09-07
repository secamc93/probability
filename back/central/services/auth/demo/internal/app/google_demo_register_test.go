package app

import (
	"context"
	"errors"
	"testing"

	"github.com/secamc93/probability/back/central/services/auth/demo/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func peticionGoogle() domain.GoogleDemoRegisterRequest {
	return domain.GoogleDemoRegisterRequest{
		SignupToken:  "token-valido",
		BusinessName: "Tienda Probability",
	}
}

func TestDemoRegisterConGoogle_CreaLaCuentaActivaYConGoogleID(t *testing.T) {
	e := nuevoEntorno(t)

	resp, err := e.uc.DemoRegisterWithGoogle(context.Background(), peticionGoogle())

	require.NoError(t, err)
	require.Len(t, e.repo.CuentasCreadas, 1)
	creada := e.repo.CuentasCreadas[0]
	assert.True(t, creada.Active, "la cuenta creada con Google no necesita verificar el correo")
	assert.Equal(t, "sub-de-prueba", creada.GoogleID)
	assert.Equal(t, "nuevo@ejemplo.com", creada.Email)
	assert.Equal(t, "Tienda Probability", creada.BusinessName)
	assert.Empty(t, creada.TokenHash, "no se emite token de verificacion por correo")
	assert.True(t, resp.Success)
	assert.Equal(t, "token-de-sesion", resp.Token)
}

func TestDemoRegisterConGoogle_NoMandaCorreoNiOTP(t *testing.T) {
	e := nuevoEntorno(t)

	_, err := e.uc.DemoRegisterWithGoogle(context.Background(), peticionGoogle())

	require.NoError(t, err)
	assert.Empty(t, e.email.Enviados)
	assert.Empty(t, e.otp.Publicados)
}

func TestDemoRegisterConGoogle_SinNombreDeNegocioFalla(t *testing.T) {
	e := nuevoEntorno(t)

	_, err := e.uc.DemoRegisterWithGoogle(context.Background(), domain.GoogleDemoRegisterRequest{
		SignupToken:  "token-valido",
		BusinessName: "   ",
	})

	assert.ErrorIs(t, err, domain.ErrBusinessNameRequired)
	assert.Empty(t, e.repo.CuentasCreadas)
}

func TestDemoRegisterConGoogle_TokenInvalidoNoCreaNada(t *testing.T) {
	e := nuevoEntorno(t)
	e.signup.ValidateFn = func(tokenString string) (*domain.GoogleSignupClaims, error) {
		return nil, errors.New("token expirado")
	}

	_, err := e.uc.DemoRegisterWithGoogle(context.Background(), peticionGoogle())

	assert.ErrorIs(t, err, domain.ErrGoogleSignupTokenInvalid)
	assert.Empty(t, e.repo.CuentasCreadas)
}

func TestDemoRegisterConGoogle_CorreoYaRegistradoNoDuplicaCuenta(t *testing.T) {
	e := nuevoEntorno(t)
	e.repo.GetDemoUserByEmailFn = func(ctx context.Context, email string) (*domain.PendingDemoUser, error) {
		return &domain.PendingDemoUser{UserID: 5, IsActive: true}, nil
	}

	_, err := e.uc.DemoRegisterWithGoogle(context.Background(), peticionGoogle())

	assert.ErrorIs(t, err, domain.ErrEmailAlreadyRegistered)
	assert.Empty(t, e.repo.CuentasCreadas)
}

func TestDemoRegisterConGoogle_LaSesionApuntaAlNegocioCreado(t *testing.T) {
	e := nuevoEntorno(t)

	resp, err := e.uc.DemoRegisterWithGoogle(context.Background(), peticionGoogle())

	require.NoError(t, err)
	require.Len(t, e.sesion.Emitidos, 1)
	emitido := e.sesion.Emitidos[0]
	assert.Equal(t, resp.UserID, emitido[0])
	assert.Equal(t, resp.BusinessID, emitido[1])
	assert.NotZero(t, resp.BusinessID)
}

func TestDemoRegisterConGoogle_NormalizaElCorreoDelPerfil(t *testing.T) {
	e := nuevoEntorno(t)
	e.signup.ValidateFn = func(tokenString string) (*domain.GoogleSignupClaims, error) {
		return &domain.GoogleSignupClaims{
			GoogleID: "sub-1",
			Email:    "  Nuevo@Ejemplo.COM ",
			Name:     "  ",
		}, nil
	}

	_, err := e.uc.DemoRegisterWithGoogle(context.Background(), peticionGoogle())

	require.NoError(t, err)
	require.Len(t, e.repo.CuentasCreadas, 1)
	assert.Equal(t, "nuevo@ejemplo.com", e.repo.CuentasCreadas[0].Email)
	assert.Equal(t, "nuevo@ejemplo.com", e.repo.CuentasCreadas[0].FullName, "sin nombre de Google se usa el correo")
}
