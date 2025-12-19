#!/bin/sh
set -e

if [ -z "$1" ]; then
  echo "usage: $0 <version>"
  echo "available: 10.0 10.1"
  exit 1
fi

VERSION="$1"

case "$VERSION" in
  10.0|10.1)
    ;;
  *)
    echo "unknown version: $VERSION"
    echo "available: 10.0 10.1"
    exit 1
    ;;
esac

mkdir -p images

IMG="images/NetBSD-$VERSION-amd64-live.img"
URL="https://cdn.netbsd.org/pub/NetBSD/images/$VERSION/NetBSD-$VERSION-amd64-live.img.gz"

if [ ! -f "$IMG" ]; then
  echo "Downloading NetBSD $VERSION live image..."
  curl -o "$IMG.gz" "$URL"
  gunzip "$IMG.gz"
fi

echo "Starting NetBSD $VERSION in QEMU..."
echo "Login as root (no password)"
echo "Then run: FTPSSLNOVERIFY=1 ftp -o - https://superarch.org/orz/install.sh | sh"
echo ""
qemu-system-x86_64 -m 512 -hda "$IMG" -net nic -net user
