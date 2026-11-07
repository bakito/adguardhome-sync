# Shared statistics and query history in two AdGuard Home panels

This community recipe describes a gateway that keeps the native AdGuard Home UI on each resolver and serves a shared read model for statistics and query history. AdGuard Home Sync continues to apply configuration from the origin to its replicas. This gateway supplies shared reads to the existing native panels. [AdGuard Home Sync](https://github.com/bakito/adguardhome-sync) · [Statistics FAQ](https://github.com/bakito/adguardhome-sync/wiki/FAQ#sync-statistics)

## Layout

| Component | Resolver A | Resolver B |
| --- | --- | --- |
| Native web API and session authority | Local AdGuard Home | Local AdGuard Home |
| DNS | Independent listener and upstream policy | Independent listener and upstream policy |
| Shared reads | Gateway, statistics merger and query index | Relay to the gateway with a labelled local fallback |
| Browser entry | Existing private HTTPS origin | Existing private HTTPS origin |

Keep the native web APIs on loopback. Serve each native UI, login, logout and settings through its existing private origin. Intercept authenticated `GET /control/stats` and `GET /control/querylog`. For each intercepted request, validate the browser session against the same node that issued it before returning shared data. The merger uses a separate restricted transport and a server-side credential. Keep session cookies bound to their respective origins. [AdGuard Home API](https://github.com/AdguardTeam/AdGuardHome/blob/master/openapi/openapi.yaml)

## Private transport example

The example uses two local listeners on A, an SSH Unix socket on A for B's native API, and a reverse TCP listener on B for the shared read endpoint. Adapt the addresses and port numbers to your deployment.

| Address | Role |
| --- | --- |
| `127.0.0.1:18053` on each node | Native AdGuard Home API |
| `127.0.0.1:18651` on each node | Browser-facing gateway behind private HTTPS |
| `127.0.0.2:18651` on A | Shared read entrance that validates B's own sessions |
| `127.0.0.1:18652` on B | Reverse SSH listener forwarding to A's shared read entrance |

Create a dedicated public-key account on B. The following OpenSSH match block restricts its TCP targets and disables shell sessions. Validate the effective configuration for this account and for the existing administrator with `sshd -T -C` before reloading SSH. Use a dedicated authorized-keys file and a pre-verified host key. [OpenSSH server configuration](https://man.openbsd.org/sshd_config)

```text
Match User adguard-panel-link
    AuthorizedKeysFile /etc/ssh/adguard-panel-link.authorized_keys
    AuthenticationMethods publickey
    PasswordAuthentication no
    KbdInteractiveAuthentication no
    AllowTcpForwarding yes
    AllowStreamLocalForwarding no
    PermitOpen 127.0.0.1:18053
    PermitListen 127.0.0.1:18652
    GatewayPorts no
    MaxSessions 0
    PermitTTY no
    PermitTunnel no
    AllowAgentForwarding no
    X11Forwarding no
Match all
```

Run a supervised SSH client on A with an owner-private runtime directory and pinned known-hosts file. `192.0.2.20` is a documentation address.

```bash
ssh -NT -F /dev/null -i /path/to/dedicated-key \
  -o BatchMode=yes -o IdentitiesOnly=yes -o IdentityAgent=none \
  -o StrictHostKeyChecking=yes -o UserKnownHostsFile=/path/to/pinned-known-hosts \
  -o GlobalKnownHostsFile=/dev/null -o ExitOnForwardFailure=yes \
  -o ServerAliveInterval=10 -o ServerAliveCountMax=2 \
  -o StreamLocalBindMask=0177 -o StreamLocalBindUnlink=yes \
  -L /path/to/private-runtime/b-native.sock:127.0.0.1:18053 \
  -R 127.0.0.1:18652:127.0.0.2:18651 adguard-panel-link@192.0.2.20
```

## Build the shared reads

1. Query each native statistics API for the requested window. Require compatible enabled retention and aligned time buckets. Add integer totals and matching chart buckets. Compute a query-count-weighted processing-time estimate. For example, two aligned buckets with 60 and 40 observed queries produce 100 observations. Check whether DNS forwarding between the nodes would make these observations unsuitable as a unique client-request count.
2. Cache each completed shared snapshot briefly so both panels use the same payload. Native top lists are truncated. Use the returned rows and omission bounds to identify a confirmed leading prefix. A broader view can show domains whose counts are known from both sources. Clearly label the ordering as limited to the displayed subset when omitted candidates may rank higher. Label incomplete lists and estimated processing times in the UI.
3. Collect native query logs into an index keyed by source and stable event identity. Make retries idempotent. For migrated snapshots, use a cutoff and source identity to avoid reimporting overlapping history. Give pagination a stable cutoff and cursor so the same page can be fetched through either panel. Display the last successful collection time and any gap.
4. Keep resets of shared history or statistics disabled through the gateways. Keep each node's configuration API and logins under that node's native authority. On transport failure, B may show local readings with a visible local-only label. A short stale snapshot needs an age and a hard expiry. DNS queries continue to use the independent resolvers.

For a synthetic ranking example, suppose both resolvers report 30 and 20 requests for the same domain. Its combined count is 50 within the reported data. If resolver B omits a second domain and its published list stops at 4 requests, the second domain may have up to 4 additional requests on B. Show the first domain's known count and keep the second count out of the exact-count view. Compare query-log coverage against the native counters before using log-derived rankings.

## Accept and roll out

1. Start the gateways on loopback while both HTTPS origins still point to the native APIs. Check both directions of the SSH link, including denial of unlisted targets.
2. Sign into each native origin separately. Check that anonymous reads and a session copied from the other origin are rejected. Verify local settings and exercise the guarded reset paths against disposable instances.
3. Compare the shared JSON from both entrances for 1 h, 24 h, 7 days and 365 days when both resolvers retain those ranges. Bracket its additive counters and series with native source reads and avoid crossing a bucket boundary. Compare a page of query history using the same cursor. Test labelled local fallback and restoration.
4. Take a private backup of each reverse proxy configuration. Switch just the existing HTTPS root on A, run the per-host checks and repeat on B. Preserve other handlers and use conditional updates when the proxy supports them. If acceptance fails, restore only the changed roots. Keep DNS on its original listeners and check A and AAAA responses over UDP and TCP before and after.
5. Confirm the dashboard and query log in separate browser sessions. The panels should display the same live snapshot for an aligned window, plus visible ranking limits and collection age. A collector delay and a relay fallback have different labels.

## Verification from a two-node deployment

The recipe was exercised with two AdGuard Home v0.107.80 instances. Separate logins worked on both native origins. The shared results matched for 1 h, 24 h, 7 days and 365 days, and the same query-log cursor returned the same page. Tests also covered anonymous access, HTTPS, DNS over UDP and TCP, and the labelled relay fallback.

## Reproduce a basic API comparison

Use two test origins configured with the gateway. Set `ADGUARD_NETRC` to a private, mode-0600 netrc file with one `machine` entry per test hostname. The requests run without interactive prompts so they can share a live cache interval. This example uses the gateway's `X-AdGuard-Preview-Status` header to verify freshness. Repeat both reads together if a snapshot boundary intervenes.

```bash
(
  set -eu
  : "${ADGUARD_NETRC:?Set the path to the private test netrc file}"
  work=$(mktemp -d)
  trap 'rm -rf -- "$work"' EXIT
  curl -fsS --netrc-file "$ADGUARD_NETRC" --dump-header "$work/a.headers" \
    'https://resolver-a.example.invalid:8443/control/stats?recent=86400000' > "$work/a.json"
  curl -fsS --netrc-file "$ADGUARD_NETRC" --dump-header "$work/b.headers" \
    'https://resolver-b.example.invalid:8443/control/stats?recent=86400000' > "$work/b.json"
  grep -Eiq '^X-AdGuard-Preview-Status: live[[:space:]]*$' "$work/a.headers"
  grep -Eiq '^X-AdGuard-Preview-Status: live[[:space:]]*$' "$work/b.headers"
  cmp "$work/a.json" "$work/b.json"
)
```

Use synthetic domains, clients and traffic in public examples. Report the windows, freshness and any ranking limitations. Keep hostnames, credentials, raw logs and infrastructure backups private.
