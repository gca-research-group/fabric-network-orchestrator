BIN="./.bin/windows_amd64"

if [[ ! -d "$BIN" ]]; then
  mkdir -p "$BIN"
fi

GOOS=windows GOARCH=amd64 go build -o "$BIN/fno.exe" ./cmd/fno
