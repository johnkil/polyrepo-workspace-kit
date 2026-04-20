#!/usr/bin/env sh
set -eu

REPO="johnkil/polyrepo-workspace-kit"
BIN_NAME="wkit"
VERSION="${WKIT_VERSION:-latest}"
INSTALL_DIR="${WKIT_INSTALL_DIR:-}"
TMP_TARGET=""

usage() {
	cat <<'EOF'
Install wkit from a GitHub Release archive.

Usage:
  sh install.sh [--version v0.2.0] [--dir /usr/local/bin]

Environment:
  WKIT_VERSION      Release tag or version. Defaults to latest.
  WKIT_INSTALL_DIR  Install directory. Defaults to the first writable PATH dir.

The installer supports macOS and Linux. It verifies checksums.txt before
installing and refuses to overwrite a symlink target.
EOF
}

fail() {
	printf 'error: %s\n' "$*" >&2
	exit 1
}

info() {
	printf '%s\n' "$*" >&2
}

while [ "$#" -gt 0 ]; do
	case "$1" in
		--version)
			shift
			[ "$#" -gt 0 ] || fail "--version requires a value"
			VERSION="$1"
			;;
		--version=*)
			VERSION="${1#--version=}"
			;;
		--dir)
			shift
			[ "$#" -gt 0 ] || fail "--dir requires a value"
			INSTALL_DIR="$1"
			;;
		--dir=*)
			INSTALL_DIR="${1#--dir=}"
			;;
		-h | --help)
			usage
			exit 0
			;;
		*)
			fail "unknown argument: $1"
			;;
	esac
	shift
done

download() {
	url="$1"
	dest="$2"
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$url" -o "$dest"
	elif command -v wget >/dev/null 2>&1; then
		wget -q "$url" -O "$dest"
	else
		fail "curl or wget is required"
	fi
}

download_stdout() {
	url="$1"
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$url"
	elif command -v wget >/dev/null 2>&1; then
		wget -qO- "$url"
	else
		fail "curl or wget is required"
	fi
}

resolve_latest_tag() {
	json="$(download_stdout "https://api.github.com/repos/${REPO}/releases/latest")"
	tag="$(printf '%s\n' "$json" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
	[ -n "$tag" ] || fail "could not resolve latest release tag"
	printf '%s\n' "$tag"
}

normalize_platform() {
	kernel="$(uname -s)"
	machine="$(uname -m)"

	case "$kernel" in
		Darwin)
			os="darwin"
			archive_ext="tar.gz"
			;;
		Linux)
			os="linux"
			archive_ext="tar.gz"
			;;
		*)
			fail "unsupported OS: $kernel"
			;;
	esac

	case "$machine" in
		x86_64 | amd64)
			arch="x86_64"
			;;
		arm64 | aarch64)
			arch="arm64"
			;;
		*)
			fail "unsupported architecture: $machine"
			;;
	esac

	printf '%s %s %s\n' "$os" "$arch" "$archive_ext"
}

first_writable_path_dir() {
	old_ifs="$IFS"
	IFS=:
	for path_dir in ${PATH:-}; do
		[ -n "$path_dir" ] || continue
		case "$path_dir" in
			/*) ;;
			*) continue ;;
		esac
		[ -d "$path_dir" ] || continue
		[ -w "$path_dir" ] || continue
		IFS="$old_ifs"
		printf '%s\n' "$path_dir"
		return 0
	done
	IFS="$old_ifs"
	return 1
}

path_contains_dir() {
	needle="$1"
	old_ifs="$IFS"
	IFS=:
	for path_dir in ${PATH:-}; do
		if [ "$path_dir" = "$needle" ]; then
			IFS="$old_ifs"
			return 0
		fi
	done
	IFS="$old_ifs"
	return 1
}

if [ "$VERSION" = "latest" ]; then
	TAG="$(resolve_latest_tag)"
else
	case "$VERSION" in
		v*) TAG="$VERSION" ;;
		*) TAG="v$VERSION" ;;
	esac
fi

VERSION_NO_V="${TAG#v}"
[ -n "$VERSION_NO_V" ] || fail "invalid version: $VERSION"

set -- $(normalize_platform)
OS="$1"
ARCH="$2"
ARCHIVE_EXT="$3"

ARCHIVE="${BIN_NAME}_${VERSION_NO_V}_${OS}_${ARCH}.${ARCHIVE_EXT}"
BASE_URL="https://github.com/${REPO}/releases/download/${TAG}"

if [ -z "$INSTALL_DIR" ]; then
	if ! INSTALL_DIR="$(first_writable_path_dir)"; then
		fail "no writable directory found in PATH; set WKIT_INSTALL_DIR to an absolute PATH directory, for example /usr/local/bin"
	fi
fi

case "$INSTALL_DIR" in
	/*) ;;
	*) fail "install directory must be absolute: $INSTALL_DIR" ;;
esac

if [ ! -d "$INSTALL_DIR" ]; then
	mkdir -p "$INSTALL_DIR" || fail "could not create install directory: $INSTALL_DIR"
fi

[ -w "$INSTALL_DIR" ] || fail "install directory is not writable: $INSTALL_DIR"

if ! path_contains_dir "$INSTALL_DIR"; then
	info "warning: $INSTALL_DIR is not in PATH"
fi

TMP_DIR="$(mktemp -d)"
cleanup() {
	rm -rf "$TMP_DIR"
	if [ -n "$TMP_TARGET" ]; then
		rm -f "$TMP_TARGET"
	fi
}
trap cleanup EXIT INT TERM

info "Downloading ${ARCHIVE}"
download "${BASE_URL}/${ARCHIVE}" "${TMP_DIR}/${ARCHIVE}"
download "${BASE_URL}/checksums.txt" "${TMP_DIR}/checksums.txt"

awk -v file="$ARCHIVE" '$2 == file { print; found = 1 } END { exit(found ? 0 : 1) }' \
	"${TMP_DIR}/checksums.txt" >"${TMP_DIR}/${ARCHIVE}.sha256" ||
	fail "checksum entry not found for ${ARCHIVE}"

if command -v sha256sum >/dev/null 2>&1; then
	(cd "$TMP_DIR" && sha256sum -c "${ARCHIVE}.sha256") >/dev/null
elif command -v shasum >/dev/null 2>&1; then
	(cd "$TMP_DIR" && shasum -a 256 -c "${ARCHIVE}.sha256") >/dev/null
else
	fail "sha256sum or shasum is required"
fi

tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "$TMP_DIR"
[ -f "${TMP_DIR}/${BIN_NAME}" ] || fail "archive did not contain ${BIN_NAME}"

TARGET="${INSTALL_DIR}/${BIN_NAME}"
if [ -L "$TARGET" ]; then
	fail "refusing to overwrite symlink: $TARGET"
fi
if [ -e "$TARGET" ] && [ ! -f "$TARGET" ]; then
	fail "refusing to overwrite non-file target: $TARGET"
fi

TMP_TARGET="${TARGET}.tmp.$$"
rm -f "$TMP_TARGET"
install -m 0755 "${TMP_DIR}/${BIN_NAME}" "$TMP_TARGET" || fail "could not stage ${TARGET}"
mv "$TMP_TARGET" "$TARGET" || {
	rm -f "$TMP_TARGET"
	fail "could not install ${TARGET}"
}
TMP_TARGET=""

info "Installed ${BIN_NAME} to ${TARGET}"
"$TARGET" version
