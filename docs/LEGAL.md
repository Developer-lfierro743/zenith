# Legal & clean-room rules

Zenith is a clean-room reimplementation. These rules are non-negotiable — they're what keeps projects like this alive (Horizon operates under the same frame).

## The rules

1. **No Mojang code.** No decompiled Minecraft source in the repo, no copied class logic. Behavior is reimplemented from protocol documentation, observable behavior, data exported from the game, and differential testing.
2. **No Mojang assets or data in the repo.** Textures, sounds, models, shipped data — none of it is committed. `cmd/zenith-prepare` builds them **on the user's machine from the user's own copy** of Minecraft: Java Edition 26.3.
3. **Never download game files.** Zenith never ships or downloads Mojang files. It reads the user's existing install (or downloads 26.3 from Mojang **for a signed-in account that owns the game**, like any third-party launcher).
4. **Ownership required.** A Microsoft account that owns Java Edition is required to play. No crack paths, ever.
5. **Prepared resources are private.** Everything `zenith-prepare` produces lands in a gitignored `.local/` directory and is never committed or shared.
6. **MIT code, trademark respect.** Code is MIT. "Minecraft" is a trademark of Mojang Synergies AB — used only nominatively, with clear "unofficial" notices.
7. **Not affiliated.** Every public surface says: not an official Minecraft product, not approved by or associated with Mojang or Microsoft.

## Reference points

- Horizon (Rust) publishes the same policy and ships the same way: <https://github.com/horizon-rust/horizon>
- Mojang EULA usage guidelines: <https://www.minecraft.net/en-us/usage-guidelines>
