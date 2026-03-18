package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := mustConfig()
	ctx := context.Background()

	pg, err := pgxpool.New(ctx, cfg.PostgresURL)
	must(err)
	defer pg.Close()

	rdb := redis.NewClient(&redis.Options{Addr: strings.TrimPrefix(cfg.RedisURL, "redis://")})
	defer rdb.Close()

	ensureGroup(ctx, rdb, cfg.RedisStream, cfg.Group)

	log.Printf("worker start stream=%s group=%s consumer=%s", cfg.RedisStream, cfg.Group, cfg.Consumer)
	lastPendingCheck := time.Now().Add(-10 * time.Minute)
	for {
		if time.Since(lastPendingCheck) > 60*time.Second {
			processPending(ctx, cfg, rdb, pg)
			lastPendingCheck = time.Now()
		}

		msgs, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    cfg.Group,
			Consumer: cfg.Consumer,
			Streams:  []string{cfg.RedisStream, ">"},
			Count:    50,
			Block:    cfg.Block,
		}).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			log.Printf("xreadgroup error: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for _, stream := range msgs {
			for _, m := range stream.Messages {
				if err := handleMessage(ctx, cfg, pg, m.Values); err != nil {
					log.Printf("handle error: %v", err)
					// don't ack -> will be retried via pending list strategy
					continue
				}
				_ = rdb.XAck(ctx, cfg.RedisStream, cfg.Group, m.ID).Err()
			}
		}
	}
}

func ensureGroup(ctx context.Context, rdb *redis.Client, stream, group string) {
	// create group at start; ignore BUSYGROUP
	_ = rdb.XGroupCreateMkStream(ctx, stream, group, "0").Err()
}

func processPending(ctx context.Context, cfg Config, rdb *redis.Client, pg *pgxpool.Pool) {
	pend, err := rdb.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: cfg.RedisStream,
		Group:  cfg.Group,
		Start:  "-",
		End:    "+",
		Count:  50,
	}).Result()
	if err != nil {
		return
	}
	if len(pend) == 0 {
		return
	}

	ids := make([]string, 0, len(pend))
	for _, p := range pend {
		// only re-claim messages idle for >=30s
		if p.Idle < 30*time.Second {
			continue
		}
		ids = append(ids, p.ID)
	}
	if len(ids) == 0 {
		return
	}

	claimed, err := rdb.XClaim(ctx, &redis.XClaimArgs{
		Stream:   cfg.RedisStream,
		Group:    cfg.Group,
		Consumer: cfg.Consumer,
		MinIdle:  30 * time.Second,
		Messages: ids,
	}).Result()
	if err != nil {
		return
	}

	for _, m := range claimed {
		if err := handleMessage(ctx, cfg, pg, m.Values); err != nil {
			continue
		}
		_ = rdb.XAck(ctx, cfg.RedisStream, cfg.Group, m.ID).Err()
	}
}

func handleMessage(ctx context.Context, cfg Config, pg *pgxpool.Pool, values map[string]any) error {
	event := fmt.Sprintf("%v", values["event"])
	payloadStr := fmt.Sprintf("%v", values["payload"])

	var payload map[string]any
	_ = json.Unmarshal([]byte(payloadStr), &payload)

	recipients := map[string]struct{}{}
	if v, ok := payload["sponsor"].(string); ok && strings.HasPrefix(v, "0x") {
		recipients[strings.ToLower(v)] = struct{}{}
	}
	if v, ok := payload["hunter"].(string); ok && strings.HasPrefix(v, "0x") {
		recipients[strings.ToLower(v)] = struct{}{}
	}
	if len(recipients) == 0 {
		return nil
	}

	for addr := range recipients {
		addrBytes, err := hex.DecodeString(strings.TrimPrefix(addr, "0x"))
		if err != nil {
			continue
		}
		_, err = pg.Exec(ctx, `
insert into notifications(user_address, kind, payload)
values($1,$2,$3)
`, addrBytes, event, []byte(payloadStr))
		if err != nil {
			return err
		}

		if err := sendWebhooks(ctx, pg, addrBytes, event, payloadStr); err != nil {
			// webhook errors are non-fatal for inbox
			log.Printf("webhook error: %v", err)
		}

		if err := sendEmailIfConfigured(ctx, cfg, pg, addrBytes, event, payloadStr); err != nil {
			log.Printf("email error: %v", err)
		}
	}

	return nil
}

func sendWebhooks(ctx context.Context, pg *pgxpool.Pool, addr []byte, event, payload string) error {
	rows, err := pg.Query(ctx, `
select url, secret
from webhooks
where user_address=$1 and enabled=true
`, addr)
	if err != nil {
		return err
	}
	defer rows.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	for rows.Next() {
		var url, secret string
		if err := rows.Scan(&url, &secret); err != nil {
			continue
		}
		sig := hmacHex(secret, payload)
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CB-Event", event)
		req.Header.Set("X-CB-Signature", sig)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		_ = resp.Body.Close()
	}
	return nil
}

func hmacHex(secret, body string) string {
	m := hmac.New(sha256.New, []byte(secret))
	_, _ = m.Write([]byte(body))
	return "sha256=" + hex.EncodeToString(m.Sum(nil))
}

func sendEmailIfConfigured(ctx context.Context, cfg Config, pg *pgxpool.Pool, addr []byte, event, payload string) error {
	if cfg.SMTPHost == "" || cfg.SMTPPort == 0 || cfg.SMTPFrom == "" {
		return nil
	}
	row := pg.QueryRow(ctx, `select email from user_emails where user_address=$1 and verified=true`, addr)
	var email string
	if err := row.Scan(&email); err != nil {
		return nil
	}

	subject := "CryptoBounty notification: " + event
	msg := "From: " + cfg.SMTPFrom + "\r\n" +
		"To: " + email + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n\r\n" +
		payload + "\r\n"

	addrPort := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	var auth smtp.Auth
	if cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)
	}
	return smtp.SendMail(addrPort, auth, cfg.SMTPFrom, []string{email}, []byte(msg))
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

