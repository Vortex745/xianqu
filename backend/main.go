package main

import (
	"embed"
	"fmt"
	"gotest/core/app"
	"io/fs"
	"os"
)

//go:embed dist/*
var content embed.FS

func main() {
	// 1. Anti-crash
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Critical Error:", err)
		}
	}()

	// 2. Setup paths
	workDir, _ := os.Getwd()
	fmt.Println(">>> Current working directory:", workDir)

	// 3. Setup Frontend FS
	frontendFS, _ := fs.Sub(content, "dist")

	// 4. Setup Engine
	r := app.SetupEngine(frontendFS)

	fmt.Println(">>> Server started successfully: http://localhost:8081")
	r.Run(":8081")
}
