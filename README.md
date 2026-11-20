# hospitium-protocol

Wire contracts between a Hospitium control plane and the devices it manages. Apache-2.0.

| Package | Contract |
|---|---|
| `v1` (`agent.go`) | Management agent ⇄ control plane: enrollment, desired state, status, one-time operations |
| `v1` (`launcher.go`) | TV launcher ⇄ control plane: pairing and guest-facing configuration |
| `v1` (`capabilities.go`) | Per-TV facts and capability observations, decided by evidence rather than platform name |

Anyone can implement either side. The reference agent is
[`hospitium-agent`](https://github.com/spectrum-labs-tech/hospitium-agent), which can also run
standalone from a local desired-state file.

## Design rules

1. **Devices only pull.** Every request goes from the device to the control plane over HTTPS. A property never needs inbound ports.
2. **Declarative for state, imperative for events.** Persistent TV configuration (`DeviceSpec`) is reconciled on every poll and must be idempotent. One-time actions (`Operation`) are separate, delivered at least once, and deduplicated by ID.
3. **Closed operation set.** There's no arbitrary shell or ADB passthrough. New operation types need a protocol change.
4. **Operations expire.** An agent must not start an operation after `expires_at`. This guard against a stale checkout cleanup wiping the next guest's sessions is the one an agent enforces on its own; a control plane adds its own checks when it delivers the operation.
5. **Guest data carries its own expiry.** The launcher must hide `GuestInfo` after `expires_at` even if it never hears from the server again.
6. **Additive evolution.** Within `v1`, fields are only ever added. Consumers ignore unknown fields. Breaking changes go into `v2` under a new path prefix.

## Agent flow

```
POST /v1/agent/enroll            enrollment_token -> agent_token (once)
loop every 30-60s ± jitter:
  GET  /v1/agent/desired-state   Bearer agent_token -> devices[] + operations[]
  ...reconcile each device over ADB, execute new operations...
  POST /v1/agent/operations/{id}/result   (per operation; repeat-safe)
  POST /v1/agent/status          observed state for every device
```

Operation results are idempotent. Reporting the same result twice has no further effect. If a result was posted for an older attempt, the server records it and does not reopen a completed operation.

## Launcher flow

```
POST /v1/launcher/pairings              -> {code, poll_token}; TV shows the code
admin claims the code in Hospitium
GET  /v1/launcher/pairings/{code}       Bearer poll_token -> device_token (once)
GET  /v1/launcher/config                Bearer device_token -> LauncherConfig (cacheable)
GET  /v1/launcher/media/{sha256}        Bearer device_token -> image (immutable)
```

### Backgrounds

`LauncherConfig.background` is the home screen's image, and `Background.Valid` is the rule for
using one:

- An image with origin `control_plane` is a media path on the control plane. Fetch it with the
  device token and cache it by address.
- An image with origin `unsplash` is an `https://images.unsplash.com` URL. Unsplash's terms require
  loading it from that URL. Fetch it with no token or cookies, and show its `attribution` whenever
  it's on screen.
- A launcher never sends its token to any other host.
- Download the `fallback_backgrounds` ahead of time. Show one when no background is assigned or the
  assigned one can't be loaded.

## Use

```bash
go get github.com/spectrum-labs-tech/hospitium-protocol@latest
```

Releases are Git tags (`vX.Y.Z`). Within `v1`, changes are additive only.

Maintainers cut a release with `task release BUMP=minor|patch` and never tag by hand. The task tags
the tip of `origin/main`, never a local commit, and only after CI has passed on that exact commit.
It also runs [`gorelease`](https://pkg.go.dev/golang.org/x/exp/cmd/gorelease) against the previous
tag and refuses anything that removes, renames or re-types part of the API. A major bump is refused
too: a breaking change is a new `v2` package. `DRY_RUN=1` runs every check without tagging.

## License

Apache-2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
