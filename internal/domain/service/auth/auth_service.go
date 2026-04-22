package auth

import (
	"SQLFactory/internal/config"
	"SQLFactory/internal/domain/entity"
	"SQLFactory/pkg/contextx"
	"SQLFactory/pkg/failure"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, v int, ttl time.Duration) error
	Get(ctx context.Context, key string) (int, error)
	Del(ctx context.Context, key string) error
	TTL(ctx context.Context, key string) time.Duration
}

type Repo interface {
	NewUser(ctx context.Context, user *entity.User) error
	CheckExist(ctx context.Context, email string) (bool, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
}

type Service struct {
	repo              Repo
	tokenCache        Cache
	emailConfirmCache Cache
	cfg               config.AuthConfig
	signingKey        []byte
}

func NewService(repo Repo, tokenCache Cache, emailConfirmCache Cache, cfg config.AuthConfig) *Service {
	return &Service{
		repo:              repo,
		tokenCache:        tokenCache,
		emailConfirmCache: emailConfirmCache,
		cfg:               cfg,
		signingKey:        []byte(cfg.SignKey),
	}
}

func (s *Service) Register(ctx context.Context, user *entity.User) error {
	const op = "auth.Register"
	l := contextx.GetLoggerOrDefault(ctx)

	exist, err := s.repo.CheckExist(ctx, user.Email)
	if err != nil {
		l.ErrorContext(ctx, "CheckExist error", "err", err)
		return fmt.Errorf("%s: %w", op, err)
	}
	if exist {
		return failure.NewAlreadyExistsError(errors.New("user already exist"))
	}

	user.Password, err = hashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := s.repo.NewUser(ctx, user); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

var ErrInvalidEmailOrPassword = errors.New("invalid email or password")

func (s *Service) Login(ctx context.Context, email string, password string) (*entity.Tokens, error) {
	const op = "auth.Login"

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if failure.IsNotFoundError(err) {
			return nil, failure.NewUnauthenticatedError(ErrInvalidEmailOrPassword)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, failure.NewUnauthenticatedError(ErrInvalidEmailOrPassword)
		}
		return nil, failure.NewInternalError(err)
	}

	accessToken, accessExpiredAt, err := s.newAccessToken(user.Id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refreshExpiredAt := time.Now().Add(s.cfg.RefreshTokenTTL)
	refresh := generateRefreshToken()
	if err := s.tokenCache.Set(ctx, refresh, user.Id, s.cfg.RefreshTokenTTL); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &entity.Tokens{
		AccessToken:           accessToken,
		RefreshToken:          refresh,
		AccessTokenExpiresAt:  accessExpiredAt,
		RefreshTokenExpiresAt: refreshExpiredAt.Unix(),
	}, nil
}

func (s *Service) Auth(_ context.Context, token string) (int, error) {
	claims, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return s.signingKey, nil
	})

	if err != nil {
		return 0, failure.NewUnauthenticatedError(errors.New("invalid token"))
	}
	mapClaims, ok := claims.Claims.(jwt.MapClaims)
	if !ok {
		return 0, failure.NewUnauthenticatedError(errors.New("invalid claims type"))
	}

	id, ok := mapClaims["id"]
	if !ok {
		return 0, failure.NewUnauthenticatedError(errors.New("id is not in token"))
	}
	tUnix, ok := mapClaims["expires"]
	if !ok {
		return 0, failure.NewUnauthenticatedError(errors.New("expires time is not in map claims"))
	}

	if time.Now().After(time.Unix(int64(tUnix.(float64)), 0)) {
		return 0, failure.NewUnauthenticatedError(errors.New("token expired"))
	}

	return int(id.(float64)), nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*entity.Tokens, error) {
	const op = "Auth.Refresh"
	userId, err := s.tokenCache.Get(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if userId == 0 {
		return nil, failure.NewUnauthenticatedError(errors.New("invalid refresh token"))
	}

	if err := s.tokenCache.Del(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	accessToken, accessExpired, err := s.newAccessToken(userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refreshExpired := time.Now().Add(s.cfg.RefreshTokenTTL)
	newRefreshToken := generateRefreshToken()
	if err := s.tokenCache.Set(ctx, newRefreshToken, userId, s.cfg.RefreshTokenTTL); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &entity.Tokens{
		AccessToken:           accessToken,
		RefreshToken:          newRefreshToken,
		AccessTokenExpiresAt:  accessExpired,
		RefreshTokenExpiresAt: refreshExpired.Unix(),
	}, nil
}

func (s *Service) newAccessToken(id int) (string, int64, error) {
	expires := time.Now().Add(s.cfg.AccessTokenTTL).Unix()

	claims := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"id":      id,
		"expires": expires,
	})
	token, err := claims.SignedString(s.signingKey)
	if err != nil {
		return "", expires, failure.NewInternalError(err)
	}
	return token, expires, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", failure.NewInternalError(err)
	}
	return string(bytes), nil
}

func generateRefreshToken() string {
	p := make([]byte, 32)
	rand.Read(p)
	return fmt.Sprintf("%x", p)
}
