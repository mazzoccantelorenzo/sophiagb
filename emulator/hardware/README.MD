# Gameboy Emulator Project Structure

## Project Structure
```
hardware/
├── cpu/              # CPU logic and instruction execution
│   └── cpu.go        # CPU struct, registers, and execution loop
│
├── memory/           # Memory management unit (MMU)
│   └── memory.go     # Memory mapping, I/O, and bank switching
│
└── assets/           # Binary assets
    └── dmg_boot.bin  # DMG boot ROM (not included)
```


## Resources

- [Pan Docs](https://gbdev.io/pandocs/) - Complete Game Boy technical reference
- [RGBDS](https://rgbds.gbdev.io/) - Game Boy assembler/linker
- [Game Boy CPU Manual](https://gbdev.io/pandocs/CPU_Instruction_Set.html)
- [SameBoy Boot Rom](https://github.com/LIJI32/SameBoy.git)