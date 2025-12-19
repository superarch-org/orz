#!/bin/sh
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
URL="https://superarch.org/orz/$OS"

if command -v curl >/dev/null; then
  curl -so orz "$URL"
elif command -v wget >/dev/null; then
  wget -qO orz "$URL"
else
  FTPSSLNOVERIFY=1 ftp -o orz "$URL"
fi

chmod +x orz
