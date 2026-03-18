package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	siwe "github.com/spruceid/siwe-go"
)

type siweVerifyRequest struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

func (s *Server) GetSIWENonce(c *fiber.Ctx) error {
	nonce, err := randomNonce()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "nonce generation failed")
	}

	// Store nonce as unused. Keyed by nonce itself for simplicity.
	key := "siwe:nonce:" + nonce
	if err := s.rdb.Set(c.Context(), key, "1", s.cfg.NonceTTL).Err(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "redis error")
	}

	return c.JSON(fiber.Map{"nonce": nonce})
}

func (s *Server) PostSIWEVerify(c *fiber.Ctx) error {
	var req siweVerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "bad json")
	}
	if req.Message == "" || req.Signature == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message/signature required")
	}

	msg, err := siwe.ParseMessage(req.Message)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid siwe message")
	}

	if msg.GetDomain() != s.cfg.SIWEDomain {
		return fiber.NewError(fiber.StatusUnauthorized, "domain mismatch")
	}
	if msg.GetURI().String() != s.cfg.SIWEURI {
		return fiber.NewError(fiber.StatusUnauthorized, "uri mismatch")
	}
	if msg.GetChainID() == nil || msg.GetChainID().Int64() != s.cfg.ChainID {
		return fiber.NewError(fiber.StatusUnauthorized, "chainId mismatch")
	}

	nonce := msg.GetNonce()
	if nonce == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing nonce")
	}

	if err := consumeNonceOnce(c.Context(), s.rdb, nonce); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "nonce invalid")
	}

	// Verify signature. siwe-go expects signature as hex string with 0x prefix.
	sig := req.Signature
	if !strings.HasPrefix(sig, "0x") {
		sig = "0x" + sig
	}
	_, err = msg.Verify(sig, &siwe.VerifyOpts{
		Domain:  s.cfg.SIWEDomain,
		Nonce:   nonce,
		Time:    time.Now(),
		ChainID: msg.GetChainID(),
	})
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "siwe verify failed")
	}

	addr := strings.ToLower(msg.GetAddress().Hex())
	token, err := s.signJWT(addr)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "jwt error")
	}
	return c.JSON(fiber.Map{"token": token, "address": addr})
}

func randomNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// URL-safe, no padding
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func consumeNonceOnce(ctx context.Context, rdb *redis.Client, nonce string) error {
	key := "siwe:nonce:" + nonce
	// GETDEL is atomic.
	val, err := rdb.GetDel(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errors.New("missing")
		}
		return err
	}
	if val != "1" {
		return errors.New("bad")
	}
	return nil
}

func (s *Server) signJWT(address string) (string, error) {
	claims := jwt.MapClaims{
		"sub": address,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Server) AuthRequired(c *fiber.Ctx) error {
	auth := c.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return fiber.NewError(fiber.StatusUnauthorized, "missing bearer token")
	}
	raw := strings.TrimPrefix(auth, "Bearer ")
	token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected alg")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid claims")
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "missing sub")
	}
	c.Locals("address", sub)
	return c.Next()
}

