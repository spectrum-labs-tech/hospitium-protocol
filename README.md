# hospitium-protocol

Wire contracts between a Hospitium control plane and the devices it manages. Apache-2.0.

| Package | Contract |
|---|---|
| `v1` (`agent.go`) | Management agent ⇄ control plane: enrollment, desired state, status, one-time operations |
| `v1` (`launcher.go`) | TV launcher ⇄ control plane: pairing and guest-facing configuration |

Anyone can implement either side. The reference agent is
[`hospitium-agent`](https://github.com/spectrum-labs-tech/hospitium-agent), which can also run
standalone from a local desired-state file.

## Design rules

1. **Devices only pull.** Every request goes from the device to the control plane over HTTPS. A property never needs inbound ports.
2. **Declarative for state, imperative for events.** Persistent TV configuration (`DeviceSpec`) is reconciled on every poll and must be idempotent. One-time actions (`Operation`) are separate, delivered at least once, and deduplicated by ID.
3. **Closed operation set.** There's no arbitrary shell or ADB passthrough. New operation types need a protocol change.
4. **Operations expire.** An agent must not start an operation after `expires_at`. This is one of two guards that stop a stale checkout cleanup from wiping the next guest's sessions. The other is server-side revalidation at delivery time.
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
```

## Use

```bash
go get github.com/spectrum-labs-tech/hospitium-protocol@latest
```

Releases are Git tags (`vX.Y.Z`). Within `v1`, changes are additive only.

## License

Apache-2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
