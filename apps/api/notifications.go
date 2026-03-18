package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type webhookRegisterRequest struct {
	URL string `json:"url"`
}

func (s *Server) GetNotifications(c *fiber.Ctx) error {
	addr := c.Locals("address").(string)
	addrBytes, err := addrToBytesLowerHex(addr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "bad address")
	}

	rows, err := s.pg.Query(c.Context(), `
select id, kind, payload, read_at, created_at
from notifications
where user_address=$1
order by created_at desc
limit 50
`, addrBytes)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}
	defer rows.Close()

	type item struct {
		ID        int64       `json:"id"`
		Kind      string      `json:"kind"`
		Payload   interface{} `json:"payload"`
		ReadAt    *string     `json:"readAt"`
		CreatedAt string      `json:"createdAt"`
	}

	var items []item
	for rows.Next() {
		var it item
		var payload []byte
		var readAt *string
		if err := rows.Scan(&it.ID, &it.Kind, &payload, &readAt, &it.CreatedAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "db scan error")
		}
		it.ReadAt = readAt
		it.Payload = jsonRaw(payload)
		items = append(items, it)
	}

	return c.JSON(fiber.Map{"items": items})
}

func (s *Server) PostWebhookRegister(c *fiber.Ctx) error {
	addr := c.Locals("address").(string)
	addrBytes, err := addrToBytesLowerHex(addr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "bad address")
	}

	var req webhookRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "bad json")
	}
	if req.URL == "" || !(strings.HasPrefix(req.URL, "https://") || strings.HasPrefix(req.URL, "http://")) {
		return fiber.NewError(fiber.StatusBadRequest, "bad url")
	}

	secret, err := randomSecret()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "secret error")
	}

	_, err = s.pg.Exec(c.Context(), `
insert into webhooks(user_address, url, secret, enabled)
values($1,$2,$3,true)
`, addrBytes, req.URL, secret)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}

	return c.JSON(fiber.Map{"ok": true})
}

func randomSecret() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// jsonRaw lets us avoid strict typing for payloads in MVP.
func jsonRaw(b []byte) fiber.Map {
	return fiber.Map{"raw": "0x" + hex.EncodeToString(b)}
}

