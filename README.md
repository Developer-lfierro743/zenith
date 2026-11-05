# Zenith

Minecraft: Java Edition 26.3, reimplemented from scratch in Go — the point directly above the horizon. 😏

Same worlds, same rules, same protocol. Built clean-room, driven by data prepared from **your own copy** of the game, and checked against vanilla with differential harnesses. Free, open source and unofficial.

> ⚠️ **Not an official Minecraft product.** Not approved by or associated with Mojang or Microsoft. Zenith contains no Minecraft code, assets or data — you need your own legally obtained copy of Minecraft: Java Edition 26.3, and a Microsoft account that owns it.

## What this is

A Go port of the *approach* behind [Horizon](https://github.com/horizon-rust/horizon) (Minecraft rewritten in Rust) — nobody had done it in Go, so here we are. Same three-piece architecture, different language:

| Piece | What it does |
|---|---|
| `cmd/zenith-launcher` | Microsoft sign-in, entitlement check, version manifest, installs/updates Zenith from your copy of the game |
| `cmd/zenith-prepare` | The real-jar trick: verifies your own 26.3 client + server jars against pinned SHA1s, runs vanilla's own data generator on a JDK to export registries (items, recipes, loot tables, biomes…), extracts assets — locally, never downloaded, never committed |
| `cmd/zenith-client` | The actual game client in Go: protocol 26.3, rendering (wgpu), input, physics |

## Status

**Pre-alpha skeleton.** Roadmap and milestones: [docs/ROADMAP.md](docs/ROADMAP.md).

- [x] M0 — repo skeleton, pinned target version
- [ ] M1 — protocol client: join a real vanilla/Fabric server
- [ ] M2 — launcher: MS auth + piston-meta
- [ ] M3 — prepare: data-gen pipeline
- [ ] M4 — rendering: chunk meshes, vanilla lighting
- [ ] M5 — playable: physics, interaction, HUD
- [ ] M6 — differential harnesses (Java oracle vs Go probe)

## Legal

Clean-room rules live in [docs/LEGAL.md](docs/LEGAL.md). Short version: MIT code, zero Mojang code/assets/data in the repo, the game never downloads or ships Mojang files, and ownership of Java Edition is required. Prepared resources are private, machine-local and gitignored.
