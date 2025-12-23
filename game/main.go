package main

// I've created this file based on the official docs of ebitengine
//
// https://ebitengine.org/en/documents/install.html
//
// This is used to create a window that displays the emulator

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct{}

type GBAFiles struct {
	Files []string
	Name  string
}

func FindGBAFiles(dir string) (GBAFiles, error) {
	var result GBAFiles

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(d.Name()) == ".gb" {
			result.Files = append(result.Files, path)
			result.Name = filepath.Base(path)
		}

		return nil
	})

	return result, err
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	gbaFiles, _ := FindGBAFiles("./roms")

	for index, file := range gbaFiles.Files {

		ebitenutil.DebugPrintAt(screen, file, 5, index+1*10)
	}
	ebitenutil.DebugPrint(screen, "Select a game to open")
	ebitenutil.OpenFile("")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	fmt.Println("test")
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
