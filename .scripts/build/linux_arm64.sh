BIN="./.bin/linux_arm64"

if [[ ! -d "$BIN" ]]; then
  mkdir -p "$BIN"
fi

GOOS=linux GOARCH=arm64 go build -o "$BIN/fno" ./cmd/fno
