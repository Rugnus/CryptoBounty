-- Minimal schema for indexer + API (shared database).

create table if not exists chain_cursors (
  chain_id bigint primary key,
  last_finalized_block bigint not null,
  last_finalized_hash bytea not null,
  updated_at timestamptz not null default now()
);

create table if not exists bounties (
  chain_id bigint not null,
  bounty_id numeric not null,
  sponsor bytea not null,
  token bytea not null,
  amount_numeric numeric not null,
  metadata_uri text not null,
  metadata_hash bytea not null,
  created_block bigint not null,
  created_tx_hash bytea not null,
  created_at timestamptz not null default now(),
  status text not null default 'Created',
  hunter bytea,
  primary key (chain_id, bounty_id)
);

create table if not exists applications (
  chain_id bigint not null,
  bounty_id numeric not null,
  hunter bytea not null,
  message_uri text not null,
  created_block bigint not null,
  created_tx_hash bytea not null,
  created_at timestamptz not null default now(),
  primary key (chain_id, bounty_id, hunter)
);

create table if not exists bounty_events (
  chain_id bigint not null,
  block_number bigint not null,
  tx_hash bytea not null,
  log_index int not null,
  bounty_id numeric,
  event_name text not null,
  payload jsonb not null,
  created_at timestamptz not null default now(),
  primary key (chain_id, block_number, tx_hash, log_index)
);

create table if not exists notifications (
  id bigserial primary key,
  user_address bytea not null,
  kind text not null,
  payload jsonb not null,
  read_at timestamptz,
  created_at timestamptz not null default now()
);

create table if not exists webhooks (
  id bigserial primary key,
  user_address bytea not null,
  url text not null,
  secret text not null,
  enabled boolean not null default true,
  created_at timestamptz not null default now()
);

create table if not exists user_emails (
  user_address bytea primary key,
  email text not null,
  verified boolean not null default false,
  created_at timestamptz not null default now()
);

