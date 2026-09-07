package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/secamc93/probability/back/central/services/auth/login/internal/domain"
	"github.com/secamc93/probability/back/central/services/auth/login/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type googleProviderStub struct {
	configurado  bool
	perfil       *domain.GoogleProfile
	err          error
	codeRecibido string
}

func (g *googleProviderStub) IsConfigured() bool { return g.configurado }

func (g *googleProviderStub) AuthCodeURL(state string) (string, error) {
	if !g.configurado {
		return "", domain.ErrGoogleNotConfigured
	}
	return "https://accounts.google.com/o/oauth2/v2/auth?state=" + state, nil
}

func (g *googleProviderStub) ExchangeCode(ctx context.Context, code string) (*domain.GoogleProfile, error) {
	g.codeRecibido = code
	if g.err != nil {
		return nil, g.err
	}
	return g.perfil, nil
}

func buildGoogleUseCase(repo *mocks.AuthRepositoryMock, provider domain.IGoogleOAuthProvider) *AuthUseCase {
	return &AuthUseCase{
		repository:  repo,
		googleOAuth: provider,
		log:         mocks.NewSilentLogger(),
		env:         &configStub{valores: map[string]string{}},
	}
}

func TestGoogleAuthURL_SinConfigurarFalla(t *testing.T) {
	uc := buildGoogleUseCase(&mocks.AuthRepositoryMock{}, &googleProviderStub{configurado: false})

	_, err := uc.GoogleAuthURL(context.Background())

	assert.ErrorIs(t, err, domain.ErrGoogleNotConfigured)
}

func TestGoogleAuthURL_GeneraStateUnico(t *testing.T) {
	uc := buildGoogleUseCase(&mocks.AuthRepositoryMock{}, &googleProviderStub{configurado: true})

	primero, err := uc.GoogleAuthURL(context.Background())
	require.NoError(t, err)
	segundo, err := uc.GoogleAuthURL(context.Background())
	require.NoError(t, err)

	assert.NotEmpty(t, primero.State)
	assert.NotEqual(t, primero.State, segundo.State)
	assert.Contains(t, primero.AuthURL, primero.State)
}

func TestLoginWithGoogle_SinCodigoFalla(t *testing.T) {
	uc := buildGoogleUseCase(&mocks.AuthRepositoryMock{}, &googleProviderStub{configurado: true})

	_, err := uc.LoginWithGoogle(context.Background(), domain.GoogleCallbackRequest{Code: "  "})

	assert.ErrorIs(t, err, domain.ErrGoogleCodeRequired)
}

func TestLoginWithGoogle_CorreoSinVerificarFalla(t *testing.T) {
	provider := &googleProviderStub{
		configurado: true,
		perfil: &domain.GoogleProfile{
			Sub:           "sub-1",
			Email:         "alguien@probabilityia.com.co",
			EmailVerified: false,
		},
	}
	uc := buildGoogleUseCase(&mocks.AuthRepositoryMock{}, provider)

	_, err := uc.LoginWithGoogle(context.Background(), domain.GoogleCallbackRequest{Code: "abc"})

	assert.ErrorIs(t, err, domain.ErrGoogleEmailNotVerified)
}

func TestLoginWithGoogle_UsuarioInexistenteNoSeCrea(t *testing.T) {
	var vinculaciones int
	repo := &mocks.AuthRepositoryMock{
		GetUserByGoogleIDFn: func(ctx context.Context, googleID string) (*domain.UserAuthInfo, error) {
			return nil, nil
		},
		GetUserByEmailFn: func(ctx context.Context, email string) (*domain.UserAuthInfo, error) {
			return nil, nil
		},
		LinkGoogleAccountFn: func(ctx context.Context, userID uint, googleID string) error {
			vinculaciones++
			return nil
		},
	}
	provider := &googleProviderStub{
		configurado: true,
		perfil: &domain.GoogleProfile{
			Sub:           "sub-1",
			Email:         "nuevo@ejemplo.com",
			EmailVerified: true,
		},
	}
	uc := buildGoogleUseCase(repo, provider)

	resultado, err := uc.LoginWithGoogle(context.Background(), domain.GoogleCallbackRequest{Code: "abc"})

	require.NoError(t, err)
	assert.True(t, resultado.NeedsSignup)
	assert.Nil(t, resultado.Session)
	assert.Equal(t, "nuevo@ejemplo.com", resultado.Profile.Email)
	assert.Zero(t, vinculaciones)
}

func TestLoginWithGoogle_BuscaPorEmailNormalizado(t *testing.T) {
	var emailConsultado string
	repo := &mocks.AuthRepositoryMock{
		GetUserByGoogleIDFn: func(ctx context.Context, googleID string) (*domain.UserAuthInfo, error) {
			return nil, nil
		},
		GetUserByEmailFn: func(ctx context.Context, email string) (*domain.UserAuthInfo, error) {
			emailConsultado = email
			return nil, nil
		},
	}
	provider := &googleProviderStub{
		configurado: true,
		perfil: &domain.GoogleProfile{
			Sub:           "sub-1",
			Email:         "  Alguien@Probabilityia.Com.Co  ",
			EmailVerified: true,
		},
	}
	uc := buildGoogleUseCase(repo, provider)

	_, _ = uc.LoginWithGoogle(context.Background(), domain.GoogleCallbackRequest{Code: "abc"})

	assert.Equal(t, "alguien@probabilityia.com.co", emailConsultado)
}

func TestLoginWithGoogle_UsuarioInactivoNoEntra(t *testing.T) {
	repo := &mocks.AuthRepositoryMock{
		GetUserByGoogleIDFn: func(ctx context.Context, googleID string) (*domain.UserAuthInfo, error) {
			return &domain.UserAuthInfo{ID: 7, Email: "x@y.com", IsActive: false}, nil
		},
		HasPendingEmailVerificationFn: func(ctx context.Context, userID uint) (bool, error) {
			return false, nil
		},
	}
	provider := &googleProviderStub{
		configurado: true,
		perfil: &domain.GoogleProfile{
			Sub:           "sub-1",
			Email:         "x@y.com",
			EmailVerified: true,
		},
	}
	uc := buildGoogleUseCase(repo, provider)

	_, err := uc.LoginWithGoogle(context.Background(), domain.GoogleCallbackRequest{Code: "abc"})

	assert.ErrorIs(t, err, domain.ErrUserInactive)
}

func TestLoginWithGoogle_VinculaCuentaExistentePorEmail(t *testing.T) {
	var vinculadoA uint
	var googleIDVinculado string
	repo := &mocks.AuthRepositoryMock{
		GetUserByGoogleIDFn: func(ctx context.Context, googleID string) (*domain.UserAuthInfo, error) {
			return nil, nil
		},
		GetUserByEmailFn: func(ctx context.Context, email string) (*domain.UserAuthInfo, error) {
			return &domain.UserAuthInfo{ID: 42, Email: email, IsActive: false}, nil
		},
		HasPendingEmailVerificationFn: func(ctx context.Context, userID uint) (bool, error) {
			return true, nil
		},
		LinkGoogleAccountFn: func(ctx context.Context, userID uint, googleID string) error {
			vinculadoA = userID
			googleIDVinculado = googleID
			return nil
		},
	}
	provider := &googleProviderStub{
		configurado: true,
		perfil: &domain.GoogleProfile{
			Sub:           "sub-99",
			Email:         "existente@ejemplo.com",
			EmailVerified: true,
		},
	}
	uc := buildGoogleUseCase(repo, provider)

	_, err := uc.LoginWithGoogle(context.Background(), domain.GoogleCallbackRequest{Code: "abc"})

	assert.ErrorIs(t, err, domain.ErrUserPendingVerification)
	assert.Equal(t, uint(42), vinculadoA)
	assert.Equal(t, "sub-99", googleIDVinculado)
}

func TestLoginWithGoogle_ErrorDelProveedorSePropaga(t *testing.T) {
	provider := &googleProviderStub{configurado: true, err: domain.ErrGoogleExchangeFailed}
	uc := buildGoogleUseCase(&mocks.AuthRepositoryMock{}, provider)

	_, err := uc.LoginWithGoogle(context.Background(), domain.GoogleCallbackRequest{Code: "abc"})

	assert.True(t, errors.Is(err, domain.ErrGoogleExchangeFailed))
}

func TestGoogleSignupToken_UsaElPerfilDeGoogle(t *testing.T) {
	var recibido []string
	jwtMock := &mocks.JWTServiceMock{
		GenerateGoogleSignupTokenFn: func(googleID, email, name, picture string, ttl time.Duration) (string, error) {
			recibido = []string{googleID, email, name, picture}
			return "token-firmado", nil
		},
	}
	uc := buildGoogleUseCase(&mocks.AuthRepositoryMock{}, &googleProviderStub{configurado: true})
	uc.jwtService = jwtMock

	token, err := uc.GoogleSignupToken(context.Background(), &domain.GoogleProfile{
		Sub:     "sub-1",
		Email:   "nuevo@ejemplo.com",
		Name:    "Nuevo Usuario",
		Picture: "https://foto",
	}, 15*time.Minute)

	require.NoError(t, err)
	assert.Equal(t, "token-firmado", token)
	assert.Equal(t, []string{"sub-1", "nuevo@ejemplo.com", "Nuevo Usuario", "https://foto"}, recibido)
}

func TestGoogleSignupToken_SinPerfilFalla(t *testing.T) {
	uc := buildGoogleUseCase(&mocks.AuthRepositoryMock{}, &googleProviderStub{configurado: true})

	_, err := uc.GoogleSignupToken(context.Background(), nil, time.Minute)

	assert.ErrorIs(t, err, domain.ErrGoogleExchangeFailed)
}
