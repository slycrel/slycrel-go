// cmd/convert_ansi — one-shot utility that re-imports the original binary
// BBS-era ANSI files from orig/.../Sly_Data/ANSI/ into data/ansi/ in the
// editable "\e[...]" format the server reads.
//
// Why: the initial port flattened CP437 box-drawing + block glyphs to ASCII
// approximations ("-", "#", "|"), losing the authentic BBS look. This tool
// walks the known CP437 bytes that actually appear in the originals (see
// byte inventory in the commit message) and maps them to their canonical
// Unicode equivalents so the re-rendered menus match the 1994 source.
//
// Usage:
//
//	go run ./cmd/convert_ansi
//
// Hard-coded paths — this is a one-off, not a reusable tool.
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// CP437 → Unicode for the ~18 high-bit glyphs actually used in the source
// files. Kept narrow on purpose: a stray 0x?? we didn't expect should fall
// through and be visible as a replacement char, not silently mapped.
var cp437 = map[byte]rune{
	0xB3: '│', 0xB4: '┤', 0xBF: '┐', 0xC0: '└',
	0xC3: '├', 0xC4: '─', 0xC5: '┼', 0xCD: '═',
	0xD8: '╪', 0xD9: '┘', 0xDA: '┌', 0xDB: '█',
	0xDC: '▄', 0xDD: '▌', 0xDE: '▐', 0xDF: '▀',
	0xF9: '∙', 0xFA: '·',
}

// Filename mapping: original binary file → target .ans filename the server
// already looks for (lower-snake-case + .ans extension).
var fileMap = map[string]string{
	"ArenaMenu":      "arena_menu.ans",
	"ArmoryMenu":     "armory_menu.ans",
	"BankMenu":       "bank_menu.ans",
	"CombatANSI":     "combat.ans",
	"HealerMenu":     "healer_menu.ans",
	"HerbalistMenu":  "herbalist_menu.ans",
	"InnMenu":        "inn_menu.ans",
	"MainMenu":       "main_menu.ans",
	"Opening ANSI":   "opening.ans",
	"TavernMenu":     "tavern_menu.ans",
	"TowerMenu":      "tower_menu.ans",
	"WildernessMenu": "wilderness_menu.ans",
}

const (
	origDir = "orig/Legacy Code/Original Slycrel Disk/Slycrel New Development/Greater Than 32K Slycrel/Sly_Data/ANSI"
	outDir  = "data/ansi"
)

func convert(input []byte) ([]byte, error) {
	var out bytes.Buffer
	// Original files mix Mac (CR only), DOS (CRLF), and Unix (LF). Walk with
	// lookahead so every line ending resolves to a single '\n'.
	for i := 0; i < len(input); i++ {
		b := input[i]
		switch {
		case b == 0x1B:
			// ESC → literal "\e". The server/terminal converts this back to
			// a real ESC byte at render time (see terminal_impl.ShowANSIFile).
			out.WriteString(`\e`)
		case b == 0x0D:
			// CR. If followed by LF, eat the pair as one newline; otherwise
			// this is a lone Mac-era CR that stands for a newline on its own.
			out.WriteByte('\n')
			if i+1 < len(input) && input[i+1] == 0x0A {
				i++
			}
		case b == 0x0A:
			out.WriteByte('\n')
		case b < 0x20 && b != 0x09:
			log.Printf("warn: dropping control byte 0x%02X", b)
		case b < 0x80:
			out.WriteByte(b)
		default:
			if r, ok := cp437[b]; ok {
				out.WriteRune(r)
			} else {
				log.Printf("warn: unmapped CP437 byte 0x%02X (emitting '?')", b)
				out.WriteByte('?')
			}
		}
	}
	return out.Bytes(), nil
}

func main() {
	converted := 0
	for src, dst := range fileMap {
		srcPath := filepath.Join(origDir, src)
		dstPath := filepath.Join(outDir, dst)

		raw, err := os.ReadFile(srcPath)
		if err != nil {
			log.Printf("skip %s: %v", src, err)
			continue
		}
		out, err := convert(raw)
		if err != nil {
			log.Fatalf("convert %s: %v", src, err)
		}
		if err := os.WriteFile(dstPath, out, 0o644); err != nil {
			log.Fatalf("write %s: %v", dstPath, err)
		}
		fmt.Printf("  %-15s → %s  (%d → %d bytes)\n", src, dst, len(raw), len(out))
		converted++
	}
	fmt.Printf("\nConverted %d files.\n", converted)
}
