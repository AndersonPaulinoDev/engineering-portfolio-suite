package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AndersonPaulinoDev/scanstock/api/internal/domain"
	"github.com/AndersonPaulinoDev/scanstock/api/internal/store"
)

type Service struct {
	secret []byte
	store  store.Store
}

type claims struct {
	Sub  string `json:"sub"`
	Role string `json:"role"`
	Exp  int64  `json:"exp"`
}

func New(secret string, st store.Store) *Service {
	return &Service{secret: []byte(secret), store: st}
}

func (s *Service) SeedDemo(ctx context.Context) error {
	users := []struct{ id, email, password, role string }{
		{"usr_stock", "stock@scanstock.dev", "demo1234", "operator"},
		{"usr_manager", "manager@scanstock.dev", "demo1234", "manager"},
	}
	for _, item := range users {
		if err := s.AddUser(ctx, item.id, item.email, item.password, item.role); err != nil && !strings.Contains(strings.ToLower(err.Error()), "exists") {
			return err
		}
	}
	return nil
}

func (s *Service) AddUser(ctx context.Context, id, email, password, role string) error {
	return s.store.CreateUser(ctx, domain.User{ID: id, Email: strings.ToLower(strings.TrimSpace(email)), PasswordHash: hash(password), Role: role})
}

func (s *Service) Login(ctx context.Context, email, password string) (domain.User, string, error) {
	user, err := s.store.UserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil || !verify(password, user.PasswordHash) {
		return domain.User{}, "", errors.New("invalid credentials")
	}
	token, err := s.sign(claims{Sub: user.ID, Role: user.Role, Exp: time.Now().Add(8 * time.Hour).Unix()})
	return user, token, err
}

func (s *Service) Verify(token string) (string, string, error) {
	var c claims
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", "", errors.New("invalid token")
	}
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(parts[0]))
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(sig, mac.Sum(nil)) {
		return "", "", errors.New("invalid token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", "", err
	}
	if err = json.Unmarshal(payload, &c); err != nil {
		return "", "", err
	}
	if time.Now().Unix() >= c.Exp {
		return "", "", errors.New("expired token")
	}
	return c.Sub, c.Role, nil
}

func (s *Service) sign(c claims) (string, error) {
	payload, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func hash(password string) string {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	derived := pbkdf2([]byte(password), salt, 80000, 32)
	return fmt.Sprintf("pbkdf2$80000$%s$%s", hex.EncodeToString(salt), hex.EncodeToString(derived))
}

func verify(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 {
		return false
	}
	iterations, _ := strconv.Atoi(parts[1])
	salt, err1 := hex.DecodeString(parts[2])
	want, err2 := hex.DecodeString(parts[3])
	if err1 != nil || err2 != nil {
		return false
	}
	return hmac.Equal(pbkdf2([]byte(password), salt, iterations, len(want)), want)
}

func pbkdf2(password, salt []byte, iter, n int) []byte {
	out := []byte{}
	for block := 1; len(out) < n; block++ {
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)
		mac.Write([]byte{byte(block >> 24), byte(block >> 16), byte(block >> 8), byte(block)})
		u := mac.Sum(nil)
		t := append([]byte(nil), u...)
		for i := 1; i < iter; i++ {
			mac = hmac.New(sha256.New, password)
			mac.Write(u)
			u = mac.Sum(nil)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		out = append(out, t...)
	}
	return out[:n]
}
