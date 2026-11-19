#!/usr/bin/env bash
# Generate a 2048-bit RSA key pair for the metrics agent and server.
#
# The private key is PKCS #8 PEM ("BEGIN PRIVATE KEY") and is meant for the
# server. The public key is PKIX PEM ("BEGIN PUBLIC KEY") and is meant for
# the agent. Both formats are what the services load by default.
#
# Usage: scripts/gen-keys.sh [private-path] [public-path]
# Defaults: private.pem and public.pem in the current directory.
# Existing files are left untouched.

set -euo pipefail

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
	cat <<'EOF'
Usage: scripts/gen-keys.sh [private-path] [public-path]

Generate a 2048-bit RSA key pair with openssl.

  private-path  PKCS #8 PEM for the server (default: private.pem)
  public-path   PKIX PEM for the agent (default: public.pem)

The script refuses to overwrite an existing file.
EOF
	exit 0
fi

private_path="${1:-private.pem}"
public_path="${2:-public.pem}"
bits=2048

if ! command -v openssl >/dev/null 2>&1; then
	echo "error: openssl is required" >&2
	exit 1
fi

for path in "$private_path" "$public_path"; do
	if [[ -e "$path" ]]; then
		echo "error: refusing to overwrite $path" >&2
		exit 1
	fi
done

mkdir -p -- "$(dirname -- "$private_path")" "$(dirname -- "$public_path")"

openssl genpkey -algorithm RSA -pkeyopt "rsa_keygen_bits:${bits}" -out "$private_path"
chmod 600 "$private_path"

openssl pkey -in "$private_path" -pubout -out "$public_path"
chmod 644 "$public_path"

echo "Server private key (PKCS #8): $private_path"
echo "Agent public key (PKIX):      $public_path"
