# Zenith roadmap

Target: Minecraft: Java Edition **26.3** (pinned in `internal/version`, mirrored by `research-lock.json`).

## M0 — Skeleton ✅
- Go module `github.com/Developer-lfierro743/zenith`
- Three binaries stubbed: `zenith-launcher`, `zenith-prepare`, `zenith-client`
- Pin file schema (`research-lock.json`), legal docs

## M1 — Protocol client joins real servers
The fastest playable win: we only own the client side; the server keeps the hard logic.
- `internal/protocol`: framing, VarInt/VarLong, compression (zlib), login → configuration → play states
- Handshake + status ping, login with MS session token, keepalive, chat
- Chunk decode: paletted sections, block states, light, entities
- Milestone test: stand in a world on a real server and see chunks render flat-colored

## M2 — Launcher
- `internal/auth`: MS OAuth device flow → XBL → XSTS → Minecraft services token; `mcstore` entitlement check (own the game or it refuses)
- `internal/meta`: piston-meta version manifest v2, version JSON parsing
- Reuse existing installs (official launcher layout, Prism/MultiMC later) or download 26.3 for a signed-in owner
- Self-update

## M3 — Prepare (the real-jar pipeline)
- Verify the user's own 26.3 client jar + server jar + asset index against `research-lock.json` SHA1 pins
- Run vanilla's own data generator on a JDK 25 to export: item components, recipes, loot tables, tags, biomes, dimension types → JSON under `.local/`
- Extract assets (textures, sounds, panorama) machine-locally; never commit, never download
- `internal/registry`: load exported JSON into data-driven registries

## M4 — Rendering
- `internal/render` on go-webgpu (same stack as the Strata rebuild): chunk streaming worker pool, section meshing with neighbour culling, smooth lighting + AO, biome tints, sky/fog
- Vanilla block models + textures from prepared assets

## M5 — Playable
- Client-side player physics (vanilla rules), mining/placing with server round-trip, inventory, HUD
- Milestone test: play on a survival server with keyboard/mouse

## M6 — Differential harnesses
- Java oracle (real game headless) vs Go probe on identical inputs; compare outputs per subsystem
- Start with: VarInt codec, chunk decode, physics stepping

## Later
- Singleplayer parity (own game logic + worldgen), Anvil save read/write, dedicated Go server, mod system
