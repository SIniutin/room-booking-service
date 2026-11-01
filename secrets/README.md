# Local development keys

PEM files in this directory are generated locally and ignored by Git.

For a local process, run:

```bash
make keys
```

Docker Compose uses a separate named volume and generates its keys automatically.
Use `make rotate-keys` to replace the keys in that volume. Rotating the keys
invalidates previously issued development tokens.

Never commit a private key. The key removed from this repository must be treated
as compromised because it remains present in old Git history; do not reuse it in
any environment.
