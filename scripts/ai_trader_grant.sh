#!/usr/bin/env bash

set -euo pipefail

DEFAULT_BINARY="./build/nuahd"
DEFAULT_CHAIN_ID="localnuah"
DEFAULT_NODE="http://localhost:26657"
DEFAULT_KEYRING_BACKEND="os"
DEFAULT_KEYRING_HOME="${HOME}/.nuahd"
DEFAULT_AUTHZ_DAYS="30"
DEFAULT_FEEGRANT_DAYS="7"

BUY_MSG_TYPE="/osmosis.assets.v1.MsgBuyAsset"
SELL_MSG_TYPE="/osmosis.assets.v1.MsgSellAsset"

print_usage() {
  cat <<'EOF'
Usage: ai_trader_grant.sh [flags]

Guides the user through granting authz/feegrant permissions to an AI trader account.
By default the script runs in dry-run mode and only prints the nuahd commands.
Pass --execute or confirm interactively to submit the transactions.

Flags:
  -b, --binary <path>            Path to nuahd binary (default: ./build/nuahd)
  -c, --chain-id <id>            Chain ID (default: localnuah)
  -n, --node <url>               RPC node address (default: http://localhost:26657)
  -k, --keyring-backend <type>   Keyring backend to use (default: os)
  -H, --keyring-home <path>      Keyring home directory (default: ~/.nuahd)
  -g, --granter-key <name>       Key name (in keyring) of the user delegating permissions
  -r, --grantee-address <addr>   Bech32 address of the AI trader / bot
  -s, --spend-limit <coins>      Spend limit for feegrant (e.g. 1000000factory/.../ndollar); leave empty to skip feegrant
      --authz-days <days|ISO>    Validity for authz grant in days or RFC3339 timestamp (default: 30)
      --feegrant-days <days|ISO> Validity for feegrant in days or RFC3339 timestamp (default: 7)
      --fees <coins>             Explicit fees to attach to each transaction (optional)
      --gas <value>              Gas value (e.g. auto, 200000) (optional)
      --gas-prices <coins>       Gas prices to use (optional)
      --execute                  Run the commands instead of just printing them
      --yes                      Skip confirmation prompt (only meaningful with --execute)
  -h, --help                     Show this message

Example:
  scripts/ai_trader_grant.sh --chain-id nuahchain-1 --granter-key alice \\
    --grantee-address nuah1xyz... --spend-limit 5000000factory/.../ndollar --execute
EOF
}

trim() {
  local trimmed="$1"
  trimmed="${trimmed#"${trimmed%%[![:space:]]*}"}"
  trimmed="${trimmed%"${trimmed##*[![:space:]]}"}"
  printf '%s' "$trimmed"
}

ensure_binary() {
  local bin="$1"
  if [[ -x "$bin" ]]; then
    printf '%s\n' "$bin"
    return
  fi

  if command -v "$bin" >/dev/null 2>&1; then
    command -v "$bin"
    return
  fi

  printf 'Error: cannot find nuahd binary at "%s"\n' "$bin" >&2
  exit 1
}

prompt_required() {
  local var_name="$1"
  local message="$2"
  local current="${!var_name:-}"

  while [[ -z "$current" ]]; do
    read -r -p "$message: " current
    current="$(trim "$current")"
  done

  printf -v "$var_name" '%s' "$current"
}

prompt_optional() {
  local var_name="$1"
  local message="$2"
  local default_value="$3"
  local current="${!var_name:-}"

  read -r -p "$message [$default_value]: " input || true
  input="$(trim "$input")"
  if [[ -z "$input" ]]; then
    input="$default_value"
  fi

  printf -v "$var_name" '%s' "$input"
}

compute_expiration() {
  local value="$1"
  local days_regex='^[0-9]+$'

  if [[ -z "$value" || "$value" == "0" ]]; then
    printf ''
    return
  fi

  if [[ "$value" =~ $days_regex ]]; then
    local days="$value"
    if command -v python3 >/dev/null 2>&1; then
      python3 - <<PY
import datetime
days = int("$days")
expiry = datetime.datetime.utcnow() + datetime.timedelta(days=days)
print(expiry.replace(microsecond=0).isoformat() + "Z")
PY
      return
    fi

    if date -u -d "+${days} days" +'%Y-%m-%dT%H:%M:%SZ' >/dev/null 2>&1; then
      date -u -d "+${days} days" +'%Y-%m-%dT%H:%M:%SZ'
      return
    fi

    if date -u -v+"${days}"d +'%Y-%m-%dT%H:%M:%SZ' >/dev/null 2>&1; then
      date -u -v+"${days}"d +'%Y-%m-%dT%H:%M:%SZ'
      return
    fi

    printf ''
    return
  fi

  printf '%s' "$value"
}

display_command() {
  printf '  '
  printf '%q ' "$@"
  printf '\n'
}

run_command() {
  local description="$1"
  shift
  echo
  echo "→ Executing: $description"
  "$@"
}

BINARY="$DEFAULT_BINARY"
CHAIN_ID="$DEFAULT_CHAIN_ID"
NODE="$DEFAULT_NODE"
KEYRING_BACKEND="$DEFAULT_KEYRING_BACKEND"
KEYRING_HOME="$DEFAULT_KEYRING_HOME"
GRANTER_KEY="${GRANTER_KEY:-}"
GRANTEE_ADDRESS="${GRANTEE_ADDRESS:-}"
SPEND_LIMIT="${SPEND_LIMIT:-}"
AUTHZ_DAYS="$DEFAULT_AUTHZ_DAYS"
FEEGRANT_DAYS="$DEFAULT_FEEGRANT_DAYS"
FEES=""
GAS_VALUE=""
GAS_PRICES=""
EXECUTE=false
AUTO_YES=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    -b|--binary)
      BINARY="$2"
      shift 2
      ;;
    -c|--chain-id)
      CHAIN_ID="$2"
      shift 2
      ;;
    -n|--node)
      NODE="$2"
      shift 2
      ;;
    -k|--keyring-backend)
      KEYRING_BACKEND="$2"
      shift 2
      ;;
    -H|--keyring-home)
      KEYRING_HOME="$2"
      shift 2
      ;;
    -g|--granter-key)
      GRANTER_KEY="$2"
      shift 2
      ;;
    -r|--grantee-address)
      GRANTEE_ADDRESS="$2"
      shift 2
      ;;
    -s|--spend-limit)
      SPEND_LIMIT="$2"
      shift 2
      ;;
    --authz-days)
      AUTHZ_DAYS="$2"
      shift 2
      ;;
    --feegrant-days)
      FEEGRANT_DAYS="$2"
      shift 2
      ;;
    --fees)
      FEES="$2"
      shift 2
      ;;
    --gas)
      GAS_VALUE="$2"
      shift 2
      ;;
    --gas-prices)
      GAS_PRICES="$2"
      shift 2
      ;;
    --execute)
      EXECUTE=true
      shift
      ;;
    --yes)
      AUTO_YES=true
      shift
      ;;
    -h|--help)
      print_usage
      exit 0
      ;;
    *)
      echo "Unknown flag: $1" >&2
      print_usage
      exit 1
      ;;
  esac
done

BINARY="$(ensure_binary "$BINARY")"

echo "📟 AI Trader Authz/Feegrant helper"
echo "Using binary: $BINARY"
echo

prompt_required GRANTER_KEY "Enter granter key name (stored in ${KEYRING_BACKEND} keyring)"
prompt_required GRANTEE_ADDRESS "Enter AI trader (grantee) bech32 address"

if [[ -z "$SPEND_LIMIT" ]]; then
  read -r -p "Enter feegrant spend limit (coins, leave empty to skip): " SPEND_LIMIT || true
  SPEND_LIMIT="$(trim "$SPEND_LIMIT")"
fi

prompt_optional AUTHZ_DAYS "Authz validity (days or RFC3339)" "$AUTHZ_DAYS"
prompt_optional FEEGRANT_DAYS "Feegrant validity (days or RFC3339, 0 to disable)" "$FEEGRANT_DAYS"

GRANTEE_ADDRESS="$(trim "$GRANTEE_ADDRESS")"
SPEND_LIMIT="$(trim "$SPEND_LIMIT")"

GRANTER_ADDRESS="$("$BINARY" keys show "$GRANTER_KEY" -a --keyring-backend "$KEYRING_BACKEND" --home "$KEYRING_HOME")"

AUTHZ_EXPIRATION="$(compute_expiration "$AUTHZ_DAYS")"
FEEGRANT_EXPIRATION="$(compute_expiration "$FEEGRANT_DAYS")"

HOME_ARGS=(--home "$KEYRING_HOME")
NODE_ARGS=()
if [[ -n "$NODE" ]]; then
  NODE_ARGS=(--node "$NODE")
fi
FEE_ARGS=()
if [[ -n "$FEES" ]]; then
  FEE_ARGS=(--fees "$FEES")
fi
GAS_ARGS=()
if [[ -n "$GAS_VALUE" ]]; then
  GAS_ARGS+=(--gas "$GAS_VALUE")
fi
if [[ -n "$GAS_PRICES" ]]; then
  GAS_ARGS+=(--gas-prices "$GAS_PRICES")
fi

AUTHZ_BASE_ARGS=(
  "$BINARY" tx authz grant "$GRANTER_ADDRESS" "$GRANTEE_ADDRESS" generic
  --chain-id "$CHAIN_ID"
  --keyring-backend "$KEYRING_BACKEND"
  --from "$GRANTER_KEY"
  "${HOME_ARGS[@]}"
  "${NODE_ARGS[@]}"
  "${FEE_ARGS[@]:-}"
  "${GAS_ARGS[@]:-}"
  --yes
)

AUTHZ_BUY_CMD=("${AUTHZ_BASE_ARGS[@]}" --msg-type="$BUY_MSG_TYPE")
AUTHZ_SELL_CMD=("${AUTHZ_BASE_ARGS[@]}" --msg-type="$SELL_MSG_TYPE")

if [[ -n "$AUTHZ_EXPIRATION" ]]; then
  AUTHZ_BUY_CMD+=("--expiration" "$AUTHZ_EXPIRATION")
  AUTHZ_SELL_CMD+=("--expiration" "$AUTHZ_EXPIRATION")
fi

FEEGRANT_CMD=()
if [[ -n "$SPEND_LIMIT" ]]; then
  FEEGRANT_CMD=(
    "$BINARY" tx feegrant grant "$GRANTER_ADDRESS" "$GRANTEE_ADDRESS"
    --spend-limit "$SPEND_LIMIT"
    --chain-id "$CHAIN_ID"
    --keyring-backend "$KEYRING_BACKEND"
    --from "$GRANTER_KEY"
    "${HOME_ARGS[@]}"
    "${NODE_ARGS[@]}"
    "${FEE_ARGS[@]:-}"
    "${GAS_ARGS[@]:-}"
    --yes
  )
  if [[ -n "$FEEGRANT_EXPIRATION" ]]; then
    FEEGRANT_CMD+=("--expiration" "$FEEGRANT_EXPIRATION")
  fi
fi

echo "Summary"
echo "-------"
echo "  Granter key    : $GRANTER_KEY"
echo "  Granter address: $GRANTER_ADDRESS"
echo "  Grantee address: $GRANTEE_ADDRESS"
echo "  Chain ID       : $CHAIN_ID"
echo "  Authz expiry   : ${AUTHZ_EXPIRATION:-<none>}"
if [[ -n "$SPEND_LIMIT" ]]; then
  echo "  Feegrant limit : $SPEND_LIMIT"
  echo "  Feegrant expiry: ${FEEGRANT_EXPIRATION:-<none>}"
else
  echo "  Feegrant       : skipped"
fi
echo
echo "The following commands will be executed:"
display_command "${AUTHZ_BUY_CMD[@]}"
display_command "${AUTHZ_SELL_CMD[@]}"
if [[ -n "$SPEND_LIMIT" ]]; then
  display_command "${FEEGRANT_CMD[@]}"
fi

if ! $EXECUTE; then
  echo
  read -r -p "Run these commands now? [y/N]: " confirmation || true
  confirmation="$(trim "$confirmation")"
  if [[ "$confirmation" =~ ^[Yy]$ ]]; then
    EXECUTE=true
  else
    echo "Dry-run complete. Copy the commands above to execute manually."
    exit 0
  fi
fi

if $EXECUTE && ! $AUTO_YES; then
  echo
  read -r -p "Final confirmation (type 'grant' to proceed): " final_check || true
  final_check="$(trim "$final_check")"
  if [[ "$final_check" != "grant" ]]; then
    echo "Aborted."
    exit 0
  fi
fi

run_command "Authz grant for MsgBuyAsset" "${AUTHZ_BUY_CMD[@]}"
run_command "Authz grant for MsgSellAsset" "${AUTHZ_SELL_CMD[@]}"

if [[ -n "$SPEND_LIMIT" ]]; then
  run_command "Feegrant allowance" "${FEEGRANT_CMD[@]}"
fi

echo
echo "✅ Done. Use 'nuahd q authz grants $GRANTER_ADDRESS $GRANTEE_ADDRESS' and 'nuahd q feegrant grant $GRANTER_ADDRESS $GRANTEE_ADDRESS' to verify."
