package service

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
	"sync"
	"time"

	"github.com/AndersonPaulinoDev/agentwatch/api/internal/domain"
	"github.com/AndersonPaulinoDev/agentwatch/api/internal/store"
)

type AuthService struct {
	store                 store.Store
	secret                []byte
	accessTTL, refreshTTL time.Duration
	mu                    sync.RWMutex
	revoked               map[string]struct{}
}
type tokenClaims struct {
	UserID string `json:"sub"`
	Role   string `json:"role"`
	Exp    int64  `json:"exp"`
	Type   string `json:"typ"`
	JTI    string `json:"jti"`
}
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func NewAuthService(st store.Store, secret string, access, refresh time.Duration) *AuthService {
	return &AuthService{store: st, secret: []byte(secret), accessTTL: access, refreshTTL: refresh, revoked: map[string]struct{}{}}
}
func (a *AuthService) Register(ctx context.Context, email, password, role string) (domain.User, TokenPair, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if !strings.Contains(email, "@") || len(password) < 8 {
		return domain.User{}, TokenPair{}, errors.New("valid email and password of at least 8 characters required")
	}
	if role == "" {
		role = "developer"
	}
	u := domain.User{ID: newID("usr"), Email: email, PasswordHash: hashPassword(password), Role: role, CreatedAt: time.Now().UTC()}
	if err := a.store.CreateUser(ctx, u); err != nil {
		return domain.User{}, TokenPair{}, err
	}
	pair, err := a.issuePair(u)
	return u, pair, err
}
func (a *AuthService) Login(ctx context.Context, email, password string) (domain.User, TokenPair, error) {
	u, err := a.store.UserByEmail(ctx, strings.TrimSpace(strings.ToLower(email)))
	if err != nil || !verifyPassword(password, u.PasswordHash) {
		return domain.User{}, TokenPair{}, errors.New("invalid credentials")
	}
	pair, err := a.issuePair(u)
	return u, pair, err
}
func (a *AuthService) Refresh(ctx context.Context, token string) (TokenPair, error) {
	c, err := a.parse(token, "refresh")
	if err != nil {
		return TokenPair{}, err
	}
	u, err := a.store.UserByID(ctx, c.UserID)
	if err != nil {
		return TokenPair{}, err
	}
	return a.issuePair(u)
}
func (a *AuthService) Logout(token string) {
	c, err := a.parseAny(token)
	if err == nil {
		a.mu.Lock()
		a.revoked[c.JTI] = struct{}{}
		a.mu.Unlock()
	}
}
func (a *AuthService) Authenticate(token string) (tokenClaims, error) {
	return a.parse(token, "access")
}
func (a *AuthService) issuePair(u domain.User) (TokenPair, error) {
	access, err := a.sign(tokenClaims{UserID: u.ID, Role: u.Role, Exp: time.Now().Add(a.accessTTL).Unix(), Type: "access", JTI: newID("tok")})
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := a.sign(tokenClaims{UserID: u.ID, Role: u.Role, Exp: time.Now().Add(a.refreshTTL).Unix(), Type: "refresh", JTI: newID("ref")})
	return TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: int64(a.accessTTL.Seconds())}, err
}
func (a *AuthService) sign(c tokenClaims) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(b)
	mac := hmac.New(sha256.New, a.secret)
	_, _ = mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payload + "." + sig, nil
}
func (a *AuthService) parse(token, typ string) (tokenClaims, error) {
	c, err := a.parseAny(token)
	if err != nil {
		return c, err
	}
	if c.Type != typ {
		return c, errors.New("wrong token type")
	}
	return c, nil
}
func (a *AuthService) parseAny(token string) (tokenClaims, error) {
	var c tokenClaims
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return c, errors.New("invalid token")
	}
	mac := hmac.New(sha256.New, a.secret)
	_, _ = mac.Write([]byte(parts[0]))
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(sig, mac.Sum(nil)) {
		return c, errors.New("invalid token")
	}
	b, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return c, err
	}
	if err = json.Unmarshal(b, &c); err != nil {
		return c, err
	}
	if time.Now().Unix() >= c.Exp {
		return c, errors.New("token expired")
	}
	a.mu.RLock()
	_, rev := a.revoked[c.JTI]
	a.mu.RUnlock()
	if rev {
		return c, errors.New("token revoked")
	}
	return c, nil
}

func hashPassword(password string) string {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	dk := pbkdf2([]byte(password), salt, 120000, 32)
	return "pbkdf2$120000$" + hex.EncodeToString(salt) + "$" + hex.EncodeToString(dk)
}
func verifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 {
		return false
	}
	iters, _ := strconv.Atoi(parts[1])
	salt, err1 := hex.DecodeString(parts[2])
	want, err2 := hex.DecodeString(parts[3])
	if err1 != nil || err2 != nil {
		return false
	}
	got := pbkdf2([]byte(password), salt, iters, len(want))
	return hmac.Equal(got, want)
}
func pbkdf2(password, salt []byte, iter, keyLen int) []byte {
	hLen := 32
	blocks := (keyLen + hLen - 1) / hLen
	out := make([]byte, 0, blocks*hLen)
	for i := 1; i <= blocks; i++ {
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)
		mac.Write([]byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)})
		u := mac.Sum(nil)
		t := append([]byte(nil), u...)
		for j := 1; j < iter; j++ {
			mac = hmac.New(sha256.New, password)
			mac.Write(u)
			u = mac.Sum(nil)
			for k := range t {
				t[k] ^= u[k]
			}
		}
		out = append(out, t...)
	}
	return out[:keyLen]
}
func newID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%x", prefix, b)
}
