BIN="./.bin/linux_amd64"

if [[ ! -d "$BIN" ]]; then
  mkdir -p "$BIN"
fi

GOOS=linux GOARCH=amd64 go build -o "$BIN/fno" ./cmd/fno
