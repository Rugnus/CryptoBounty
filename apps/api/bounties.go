package main

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type bountyRow struct {
	ChainID      int64  `json:"chainId"`
	BountyID     string `json:"bountyId"`
	Sponsor      string `json:"sponsor"`
	Token        string `json:"token"`
	Amount       string `json:"amount"`
	MetadataURI  string `json:"metadataUri"`
	MetadataHash string `json:"metadataHash"`
	Status       string `json:"status"`
	Hunter       string `json:"hunter,omitempty"`
}

func (s *Server) GetBounties(c *fiber.Ctx) error {
	// Minimal search: q matches metadata_uri (in MVP), status filter, pagination.
	q := c.Query("q")
	status := c.Query("status")
	limit := clampInt(parseIntDefault(c.Query("limit"), 20), 1, 50)
	offset := clampInt(parseIntDefault(c.Query("offset"), 0), 0, 10_000)

	where := []string{"chain_id = $1"}
	args := []any{s.cfg.ChainID}
	i := 2

	if status != "" {
		where = append(where, "status = $"+strconv.Itoa(i))
		args = append(args, status)
		i++
	}
	if q != "" {
		where = append(where, "metadata_uri ilike $"+strconv.Itoa(i))
		args = append(args, "%"+q+"%")
		i++
	}

	sql := `
select chain_id, bounty_id::text, sponsor, token, amount_numeric::text, metadata_uri, metadata_hash, status, hunter
from bounties
where ` + strings.Join(where, " and ") + `
order by created_block desc
limit $` + strconv.Itoa(i) + ` offset $` + strconv.Itoa(i+1)

	args = append(args, limit, offset)

	rows, err := s.pg.Query(c.Context(), sql, args...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}
	defer rows.Close()

	out := []bountyRow{}
	for rows.Next() {
		var r bountyRow
		var sponsor, token, metadataHash, hunter []byte
		if err := rows.Scan(&r.ChainID, &r.BountyID, &sponsor, &token, &r.Amount, &r.MetadataURI, &metadataHash, &r.Status, &hunter); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "db scan error")
		}
		r.Sponsor = "0x" + hex.EncodeToString(sponsor)
		r.Token = "0x" + hex.EncodeToString(token)
		r.MetadataHash = "0x" + hex.EncodeToString(metadataHash)
		if len(hunter) == 20 {
			r.Hunter = "0x" + hex.EncodeToString(hunter)
		}
		out = append(out, r)
	}
	return c.JSON(fiber.Map{"items": out, "limit": limit, "offset": offset})
}

func (s *Server) GetBountyByID(c *fiber.Ctx) error {
	id := c.Params("id")
	row := s.pg.QueryRow(c.Context(), `
select chain_id, bounty_id::text, sponsor, token, amount_numeric::text, metadata_uri, metadata_hash, status, hunter
from bounties
where chain_id=$1 and bounty_id=$2
`, s.cfg.ChainID, id)

	var r bountyRow
	var sponsor, token, metadataHash, hunter []byte
	if err := row.Scan(&r.ChainID, &r.BountyID, &sponsor, &token, &r.Amount, &r.MetadataURI, &metadataHash, &r.Status, &hunter); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "not found")
	}
	r.Sponsor = "0x" + hex.EncodeToString(sponsor)
	r.Token = "0x" + hex.EncodeToString(token)
	r.MetadataHash = "0x" + hex.EncodeToString(metadataHash)
	if len(hunter) == 20 {
		r.Hunter = "0x" + hex.EncodeToString(hunter)
	}
	return c.JSON(r)
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

