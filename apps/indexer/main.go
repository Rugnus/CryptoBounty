package main

import (
	"context"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

//go:embed abi_bounty_escrow.json
var escrowABIJSON string

func main() {
	cfg := mustConfig()
	ctx := context.Background()

	pg, err := pgxpool.New(ctx, cfg.PostgresURL)
	must(err)
	defer pg.Close()
	must(migrate(ctx, pg))

	rdb := redis.NewClient(&redis.Options{Addr: strings.TrimPrefix(cfg.RedisURL, "redis://")})
	defer rdb.Close()

	client, err := ethclient.Dial(cfg.RPCURL)
	must(err)

	contractABI, err := abi.JSON(strings.NewReader(escrowABIJSON))
	must(err)

	escrowAddr := common.HexToAddress(cfg.EscrowAddress)
	eventIDs := map[common.Hash]string{}
	for _, ev := range contractABI.Events {
		eventIDs[ev.ID] = ev.Name
	}

	log.Printf("indexer start chain=%d escrow=%s conf=%d", cfg.ChainID, escrowAddr, cfg.Confirmations)

	// init cursor
	head, err := client.BlockNumber(ctx)
	must(err)
	start := uint64(0)
	if head > cfg.BackfillBlocksOnStart {
		start = head - cfg.BackfillBlocksOnStart
	}
	cursor, cursorHash, ok := loadCursor(ctx, pg, cfg.ChainID)
	if ok {
		// re-verify last finalized block hash to detect reorg
		h, err := client.HeaderByNumber(ctx, big.NewInt(cursor))
		if err == nil && h != nil {
			if !bytesEq(cursorHash, h.Hash().Bytes()) {
				log.Printf("reorg detected at cursor=%d, resetting backfill window", cursor)
				start = uint64(maxI64(0, cursor-int64(cfg.BackfillBlocksOnStart)))
				if err := resetFromBlock(ctx, pg, cfg.ChainID, int64(start)); err != nil {
					log.Printf("reset error: %v", err)
				}
			} else {
				start = uint64(cursor + 1)
			}
		}
	}

	lastProcessed := start
	for {
		head, err = client.BlockNumber(ctx)
		if err != nil {
			log.Printf("rpc head error: %v", err)
			time.Sleep(cfg.PollInterval)
			continue
		}
		if head < cfg.Confirmations {
			time.Sleep(cfg.PollInterval)
			continue
		}
		finalized := head - cfg.Confirmations
		if lastProcessed > finalized {
			time.Sleep(cfg.PollInterval)
			continue
		}

		from := lastProcessed
		to := minU64(finalized, from+2000) // batch
		if err := indexRange(ctx, client, pg, rdb, cfg, contractABI, eventIDs, escrowAddr, from, to); err != nil {
			log.Printf("index range error: %v", err)
			time.Sleep(cfg.PollInterval)
			continue
		}
		lastProcessed = to + 1
	}
}

func indexRange(
	ctx context.Context,
	client *ethclient.Client,
	pg *pgxpool.Pool,
	rdb *redis.Client,
	cfg Config,
	contractABI abi.ABI,
	eventIDs map[common.Hash]string,
	escrowAddr common.Address,
	from, to uint64,
) error {
	q := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(from),
		ToBlock:   new(big.Int).SetUint64(to),
		Addresses: []common.Address{escrowAddr},
	}
	logs, err := client.FilterLogs(ctx, q)
	if err != nil {
		return err
	}

	for _, lg := range logs {
		if len(lg.Topics) == 0 {
			continue
		}
		name, ok := eventIDs[lg.Topics[0]]
		if !ok {
			continue
		}

		payload, bountyID, err := decodeEvent(contractABI, name, lg)
		if err != nil {
			return fmt.Errorf("decode %s: %w", name, err)
		}

		if err := upsertEventAndProject(ctx, pg, cfg.ChainID, name, lg, bountyID, payload); err != nil {
			return err
		}

		// emit to Redis stream for notifications/worker
		_, _ = rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: cfg.RedisStream,
			Values: map[string]any{
				"chainId":     cfg.ChainID,
				"event":       name,
				"blockNumber": lg.BlockNumber,
				"txHash":      lg.TxHash.Hex(),
				"logIndex":    lg.Index,
				"bountyId":    bountyID.String(),
				"payload":     string(payload),
			},
		}).Result()
	}

	// update cursor at `to`
	h, err := client.HeaderByNumber(ctx, new(big.Int).SetUint64(to))
	if err != nil {
		return err
	}
	return saveCursor(ctx, pg, cfg.ChainID, int64(to), h.Hash().Bytes())
}

func decodeEvent(contractABI abi.ABI, name string, lg types.Log) (payload []byte, bountyID *big.Int, err error) {
	ev := contractABI.Events[name]

	out := map[string]any{}

	// --- Разделяем аргументы на indexed и non-indexed вручную ---
	var indexedArgs []abi.Argument
	var nonIndexedArgs abi.Arguments

	for _, arg := range ev.Inputs {
		if arg.Indexed {
			indexedArgs = append(indexedArgs, arg)
		} else {
			nonIndexedArgs = append(nonIndexedArgs, arg)
		}
	}

	// decode non-indexed
	if len(lg.Data) > 0 {
		m := map[string]any{}
		if err := nonIndexedArgs.UnpackIntoMap(m, lg.Data); err != nil {
			return nil, nil, err
		}
		for k, v := range m {
			out[k] = normalizeABIValue(v)
		}
	}

	// decode indexed
	for i, arg := range indexedArgs {
		if len(lg.Topics) <= i+1 {
			continue
		}
		topic := lg.Topics[i+1]
		switch arg.Type.T {
		case abi.AddressTy:
			out[arg.Name] = common.BytesToAddress(topic.Bytes()).Hex()
		case abi.UintTy, abi.IntTy:
			out[arg.Name] = new(big.Int).SetBytes(topic.Bytes()).String()
		case abi.FixedBytesTy:
			out[arg.Name] = "0x" + hex.EncodeToString(topic.Bytes())
		default:
			out[arg.Name] = "0x" + hex.EncodeToString(topic.Bytes())
		}
	}

	if v, ok := out["bountyId"]; ok {
		switch t := v.(type) {
		case string:
			bountyID = new(big.Int)
			_, _ = bountyID.SetString(t, 10)
		}
	}
	if bountyID == nil {
		bountyID = big.NewInt(0)
	}

	payload, err = json.Marshal(out)
	return payload, bountyID, err
}

func normalizeABIValue(v any) any {
	switch t := v.(type) {
	case common.Address:
		return t.Hex()
	case common.Hash:
		return t.Hex()
	case *big.Int:
		return t.String()
	case []byte:
		return "0x" + hex.EncodeToString(t)
	case [32]byte:
		return "0x" + hex.EncodeToString(t[:])
	case [20]byte:
		return "0x" + hex.EncodeToString(t[:])
	default:
		return v
	}
}

func toHexString(v any) string {
	switch t := v.(type) {
	case string:
		// Ensure even number of hex digits
		s := strings.TrimPrefix(t, "0x")
		if len(s)%2 != 0 {
			s = "0" + s
		}
		return "0x" + s
	case []interface{}:
		// JSON-десериализованный байт-массив [171, 205, ...]
		b := make([]byte, len(t))
		for i, x := range t {
			if f, ok := x.(float64); ok {
				b[i] = byte(f)
			}
		}
		return "0x" + hex.EncodeToString(b)
	default:
		return fmt.Sprintf("0x%x", v)
	}
}

func upsertEventAndProject(
	ctx context.Context,
	pg *pgxpool.Pool,
	chainID int64,
	eventName string,
	lg types.Log,
	bountyID *big.Int,
	payload []byte,
) error {
	tx, err := pg.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
insert into bounty_events(chain_id, block_number, tx_hash, log_index, bounty_id, event_name, payload)
values($1,$2,$3,$4,$5,$6,$7)
on conflict do nothing
`, chainID, int64(lg.BlockNumber), lg.TxHash.Hex(), int32(lg.Index), bountyID.String(), eventName, payload)
	if err != nil {
		return err
	}

	// lightweight projections for catalog
	switch eventName {
	case "BountyCreated":
		var m map[string]any
		_ = json.Unmarshal(payload, &m)
		amount := fmt.Sprintf("%v", m["amount"])
		metadataURI := fmt.Sprintf("%v", m["metadataURI"])
		metadataHash := toHexString(m["metadataHash"]) // <-- FIX
		sponsor := toHexString(m["sponsor"])           // <-- FIX (на всякий случай)
		token := toHexString(m["token"])               // <-- FIX (на всякий случай)
		_, err = tx.Exec(ctx, `
insert into bounties(chain_id, bounty_id, sponsor, token, amount_numeric, metadata_uri, metadata_hash, created_block, created_tx_hash, status)
values($1,$2,decode(trim(leading '0x' from $3),'hex'),decode(trim(leading '0x' from $4),'hex'),$5,$6,decode(trim(leading '0x' from $7),'hex'),$8,$9,'Created')
on conflict (chain_id, bounty_id) do update set
  sponsor=excluded.sponsor,
  token=excluded.token,
  amount_numeric=excluded.amount_numeric,
  metadata_uri=excluded.metadata_uri,
  metadata_hash=excluded.metadata_hash,
  status='Created'
`, chainID, bountyID.String(), sponsor, token, amount, metadataURI, metadataHash, int64(lg.BlockNumber), lg.TxHash.Hex())
	case "ApplicationSubmitted":
		var m map[string]any
		_ = json.Unmarshal(payload, &m)
		hunter := toHexString(m["hunter"]) // <-- FIX
		messageURI := fmt.Sprintf("%v", m["messageURI"])
		_, err = tx.Exec(ctx, `
insert into applications(chain_id, bounty_id, hunter, message_uri, created_block, created_tx_hash)
values($1,$2,decode(trim(leading '0x' from $3),'hex'),$4,$5,$6)
on conflict do update set message_uri=excluded.message_uri
`, chainID, bountyID.String(), hunter, messageURI, int64(lg.BlockNumber), lg.TxHash.Hex())
	case "HunterAssigned":
		var m map[string]any
		_ = json.Unmarshal(payload, &m)
		hunter := toHexString(m["hunter"]) // <-- FIX
		_, err = tx.Exec(ctx, `
update bounties set status='Assigned', hunter=decode(trim(leading '0x' from $3),'hex')
where chain_id=$1 and bounty_id=$2
`, chainID, bountyID.String(), hunter)
	case "WorkSubmitted":
		_, err = tx.Exec(ctx, `
update bounties set status='Submitted' where chain_id=$1 and bounty_id=$2
`, chainID, bountyID.String())
	case "Approved":
		_, err = tx.Exec(ctx, `
update bounties set status='Approved' where chain_id=$1 and bounty_id=$2
`, chainID, bountyID.String())
	case "Disputed":
		_, err = tx.Exec(ctx, `
update bounties set status='Disputed' where chain_id=$1 and bounty_id=$2
`, chainID, bountyID.String())
	case "PaidOut":
		_, err = tx.Exec(ctx, `
update bounties set status='PaidOut' where chain_id=$1 and bounty_id=$2
`, chainID, bountyID.String())
	case "Refunded":
		_, err = tx.Exec(ctx, `
update bounties set status='Refunded' where chain_id=$1 and bounty_id=$2
`, chainID, bountyID.String())
	}
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func resetFromBlock(ctx context.Context, pg *pgxpool.Pool, chainID int64, fromBlock int64) error {
	// best-effort cleanup so reorged logs are re-projected deterministically
	tx, err := pg.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `delete from bounty_events where chain_id=$1 and block_number >= $2`, chainID, fromBlock)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `delete from applications where chain_id=$1 and created_block >= $2`, chainID, fromBlock)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `delete from bounties where chain_id=$1 and created_block >= $2`, chainID, fromBlock)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func loadCursor(ctx context.Context, pg *pgxpool.Pool, chainID int64) (block int64, hash []byte, ok bool) {
	row := pg.QueryRow(ctx, `select last_finalized_block, last_finalized_hash from chain_cursors where chain_id=$1`, chainID)
	if err := row.Scan(&block, &hash); err != nil {
		return 0, nil, false
	}
	return block, hash, true
}

func saveCursor(ctx context.Context, pg *pgxpool.Pool, chainID, block int64, hash []byte) error {
	_, err := pg.Exec(ctx, `
insert into chain_cursors(chain_id, last_finalized_block, last_finalized_hash)
values($1,$2,$3)
on conflict (chain_id) do update set
  last_finalized_block=excluded.last_finalized_block,
  last_finalized_hash=excluded.last_finalized_hash,
  updated_at=now()
`, chainID, block, hash)
	return err
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func bytesEq(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func minU64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func maxI64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
