//go:generate goversioninfo -icon=resources/windows/dxnetclient.ico -manifest=resources/windows/dxnetclient.exe.manifest.xml -64=true -o=dxnetclient.syso

/*
Copyright © 2022 Netmaker Team <info@netmaker.io>
*/
package main

import (
	"github.com/gravitl/netclient/cmd"
	"github.com/gravitl/netclient/config"
)

// TODO: use -ldflags to set the right version at build time
var version = "v0.21.0"

func main() {
	config.SetVersion(version)
	cmd.Execute()
}
