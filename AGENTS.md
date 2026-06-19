# AGENTS.md

This file provides guidance to coding agents (e.g. Claude Code, claude.ai/code) when working with code in this repository.

## Repository purpose

Go module `kubestash.dev/apimachinery` — the canonical API types, CRDs, generated DeepCopy, admission-webhook handlers, and shared utility packages for [KubeStash](https://kubestash.com/), AppsCode's Kubernetes-native backup/restore platform. Library only; downstream binaries (`kubestash` operator, CLI, addon binaries) import these types.

Three API groups (under domain `kubestash.com`):
- `core.kubestash.com` — `BackupConfiguration`, `BackupBatch`, `BackupBlueprint`, `BackupSession`, `BackupVerifier`, `BackupVerificationSession`, `HookTemplate`, `RestoreSession`.
- `addons.kubestash.com` — `Addon`, `Function` (the per-engine backup/restore plug-ins).
- `storage.kubestash.com` — storage-side types (`BackupStorage`, `Repository`, `Snapshot`, `RetentionPolicy`).
- `config.kubestash.com/v1alpha1` — runtime config types.

## Architecture

- `apis/`:
  - Each group lives under `apis/<group>/v1alpha1/` with `*_types.go`, hand-written helpers, and generated `zz_generated.deepcopy.go`.
  - `apis/<group>/install/` and `apis/<group>/fuzzer/` — standard k8s scheme registration and round-trip helpers.
  - Top-level `apis/constant.go`, `apis/types.go`, `apis/variables.go`, `apis/helpers.go`, `apis/zz_generated.deepcopy.go` — cross-group shared types and constants.
- `crds/` — generated CRD YAML manifests for all groups, written under `kubestash.com` domain (e.g. `core.kubestash.com_backupconfigurations.yaml`).
- `webhooks/`:
  - `core/`, `storage/` — admission webhook handlers (validating + defaulting) for the respective CRDs. Live with the apimachinery, not with the operator, because they're imported by both.
- `pkg/`:
  - `blob/` — blob-store abstraction used to write/read snapshots.
  - `cloud/` — cloud provider plumbing (AWS, Azure, GCP, etc.) consumed by `blob/`.
  - `snapshot/` — snapshot read/write, including `pkg/snapshot/io/` for storage IO.
  - `resolver/` — resource resolver (turns a target ref into a concrete object set).
  - `resourceops/` — Kubernetes resource operations (delete-with-grace, wait helpers).
  - `retry/` — retry helpers used across operator and addon.
  - `workerpool/` — worker pool used by parallel backup/restore.
  - `version.go`, `utils.go`, `version_test.go` — module-level helpers.
- `PROJECT` — Kubebuilder project metadata. Domain is `kubestash.com`, **multigroup = true**.
- `hack/` — codegen scripts; `Makefile` is a Kubebuilder-style harness with a **local Go toolchain**. Tools (`controller-gen`, `kustomize`, `envtest`) install into `bin/`.
- `vendor/` — checked-in deps.

## Common commands

This repo uses a **local Go toolchain** (Kubebuilder Makefile pattern), not the AppsCode Docker harness.

- `make build` (alias `make all`) — `manifests generate fmt vet`, then build.
- `make generate` — controller-gen DeepCopy generation.
- `make manifests` — controller-gen CRDs / RBAC / webhook manifests.
- `make label-crds` — apply standard labels to generated CRDs.
- `make fmt` — `go fmt ./...`.
- `make vet` — `go vet ./...`.
- `make test` — `manifests generate fmt vet envtest`, then Go tests (uses `setup-envtest` for controller-runtime tests).
- `make run` — run a controller against `~/.kube/config` (mostly for sanity; this is a lib repo).
- `make docker-build` — `test`, then docker build.
- `make docker-push` — push the built image.
- `make docker-buildx` — multi-arch build+push.
- `make install` / `make uninstall` — `kustomize` apply/remove CRDs against the current kube context.
- `make help` — list all targets.

Run a single Go test:

```
go test ./apis/core/v1alpha1/... -run TestName -v
```

Webhook tests (using envtest):

```
go test ./webhooks/core/... -run TestName -v
```

## Conventions

- Module path is `kubestash.dev/apimachinery` (vanity URL). Imports must use that.
- Domain is `kubestash.com`. CRD groups are `core.kubestash.com`, `addons.kubestash.com`, `storage.kubestash.com`, `config.kubestash.com`. Do not change these without coordinating with every downstream KubeStash repo.
- License: see `LICENSE` (AppsCode); add the standard header to new files.
- Sign off commits (`git commit -s`); contributions follow the project's DCO requirement.
- Vendor directory is checked in — keep `go mod tidy && go mod vendor` clean.
- Do not hand-edit `zz_generated.deepcopy.go` or anything under `crds/` — change `apis/<group>/v1alpha1/*_types.go` and re-run `make manifests generate`.
- This is a **Kubebuilder multigroup project** (`PROJECT` `multigroup: true`). Use `kubebuilder` to scaffold new APIs; don't hand-create files that `PROJECT` should track.
- Webhooks live with the types (`webhooks/<group>/`) — operators import this package, register the webhook handlers, and don't redefine validation/defaulting elsewhere.
- New cloud blob backend goes under `pkg/cloud/<provider>/`; surface it through `pkg/blob/`, not by branching across `pkg/snapshot/`.
