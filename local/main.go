package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/4Kaze/birthdaybot/notifier/adapters"
)

// Example usage: go run main.go path/to/notifier/resources example.webp
func main() {
	if len(os.Args) < 3 {
		log.Fatalln("Please provide a path to resources directory and image")
	}
	resourcesDir, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	imagePath, err := filepath.Abs(os.Args[2])
	if err != nil {
		log.Fatalln(err)
	}
	videoGenerator := adapters.NewVideoGenerator(resourcesDir)
	videoGenerator.CreateVideo(imagePath)
}
