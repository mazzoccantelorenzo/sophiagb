gameboy-emulator/
├── cmd/
│   └── emulator/
│       └── main.go              # Entry point, initialize and run emulator
│
├── pkg/
│   ├── cpu/
│   │   ├── cpu.go               # CPU struct and base methods
│   │   ├── instructions.go      # Instructions implementation (opcode 0x00-0xFF)
│   │   ├── instructions_cb.go   # CB-prefixed instructions (0xCB XX)
│   │   └── registers.go         # Register helpers (getHL, setBC, etc.)
│   │
│   ├── memory/
│   │   ├── memory.go            # Memory struct and Read/Write
│   │   ├── mmu.go               # Memory Management Unit (address mapping)
│   │   └── cartridge.go         # ROM loading, MBC (Memory Bank Controllers)
│   │
│   ├── ppu/
│   │   ├── ppu.go               # Picture Processing Unit main
│   │   ├── renderer.go          # Scanline, tiles, sprites rendering
│   │   ├── lcd.go               # LCD registers (LCDC, STAT, etc.)
│   │   └── palette.go           # Color palette management
│   │
│   ├── input/
│   │   └── joypad.go            # Input (A, B, Start, Select, D-pad)
│   │
│   ├── timer/
│   │   └── timer.go             # Timer and Divider registers
│   │
│   ├── interrupts/
│   │   └── interrupts.go        # Interrupt system (VBlank, Timer, etc.)
│   │
│   └── gameboy/
│       └── gameboy.go           # Main struct that connects everything
│
├── internal/
│   └── debug/
│       ├── debugger.go          # Debugger (optional, very useful!)
│       └── logger.go            # Instruction logging
│
├── test/
│   ├── roms/                    # Test ROMs (cpu_instrs, etc.)
│   └── cpu_test.go              # Unit tests for instructions
│
├── assets/
│   └── boot_rom/
│       └── dmg_boot.bin         # Boot ROM (optional)
│
├── go.mod
├── go.sum
└── README.md
