package main

import "github.com/StemSplit/stemsplit-cli/cmd"

// version is set at build time via -ldflags "-X main.version=x.y.z"
var version = "dev"

func main() {
	cmd.Execute(version)
}
