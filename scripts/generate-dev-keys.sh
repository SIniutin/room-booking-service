#!/bin/sh

set -eu

key_dir="${1:-secrets}"
private_key="${key_dir}/private.pem"
public_key="${key_dir}/public.pem"

if [ "${FORCE_REGENERATE:-0}" != "1" ] && [ -s "${private_key}" ] && [ -s "${public_key}" ]; then
    echo "JWT development keys already exist in ${key_dir}"
    exit 0
fi

mkdir -p "${key_dir}"
tmp_dir="$(mktemp -d "${key_dir}/.jwt-keys.XXXXXX")"
trap 'rm -rf "${tmp_dir}"' EXIT INT TERM

umask 077
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out "${tmp_dir}/private.pem"
openssl pkey -in "${tmp_dir}/private.pem" -pubout -out "${tmp_dir}/public.pem"
chmod 600 "${tmp_dir}/private.pem"
chmod 644 "${tmp_dir}/public.pem"

mv "${tmp_dir}/private.pem" "${private_key}"
mv "${tmp_dir}/public.pem" "${public_key}"

echo "Generated fresh JWT development keys in ${key_dir}"
