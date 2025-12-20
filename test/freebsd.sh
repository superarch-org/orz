#!/bin/sh
set -e

if [ -z "$1" ]; then
  echo "usage: $0 <version>"
  echo "available: 13.5 14.3 15.0"
  exit 1
fi

VERSION="$1"

case "$VERSION" in
  13.5|14.3|15.0)
    ;;
  *)
    echo "unknown version: $VERSION"
    echo "available: 13.5 14.3 15.0"
    exit 1
    ;;
esac

mkdir -p images

IMG="images/FreeBSD-$VERSION-RELEASE-amd64-disc1.iso"
URL="https://download.freebsd.org/releases/amd64/amd64/ISO-IMAGES/$VERSION/FreeBSD-$VERSION-RELEASE-amd64-disc1.iso"

if [ ! -f "$IMG" ]; then
  echo "Downloading FreeBSD $VERSION..."
  curl -o "$IMG" "$URL"
fi

# Boot it
echo "Starting FreeBSD $VERSION in QEMU..."
echo "Select 'Shell' from installer menu to get a live environment"
echo "Then run:"
echo "  dhclient em0"
echo "  cd /tmp"
echo "  fetch -o orz https://superarch.org/orz/freebsd"
echo "  chmod +x orz"
echo "  HOME=/tmp ./orz run ncdu"
echo ""
qemu-system-x86_64 -m 512 -cdrom "$IMG" -net nic -net user
