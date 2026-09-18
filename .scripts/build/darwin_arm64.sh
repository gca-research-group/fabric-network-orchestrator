BIN="./.bin/darwin_arm64"

if [[ ! -d "$BIN" ]]; then
  mkdir -p "$BIN"
fi

GOOS=darwin GOARCH=arm64 go build -o "$BIN/fno" ./cmd/fno
