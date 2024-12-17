/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mlayerprotocol/go-mlayer/cmd"
)
var version string
var releaseDate string

func main() {
	if version == "" {
		version = "v1.0.0"
	}
	if releaseDate == "" {
		releaseDate = "Released: October 5th 2024 03:23 UTC"
	}
	err := os.Setenv("CLIENT_VERSION", version)
	if err != nil {
		fmt.Println("Error setting client version:", err)
	}
	err = os.Setenv("RELEASE_DATE", releaseDate)
	if err != nil {
		fmt.Println("Error setting release date:", err)
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	cmd.Execute(version, releaseDate)
}
