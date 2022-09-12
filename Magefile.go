//+build mage

package main

import (
	"fmt"
	// mage:import
	build "github.com/grafana/grafana-plugin-sdk-go/build"
	"github.com/magefile/mage/mg"
)

// Hello prints a message (shows that you can define custom Mage targets).
func Hello() {
	fmt.Println("Building db-2-datasource!")
}

// Only build for windows, because our ibm_db_go import is built for windows only.
// Need to build for only Linux on a Linux platform, with ibm_db_go installed for Linux in order to build for Linux.
func BuildWin() {
	b := build.Build{}
	mg.Deps(b.Windows)
}

var Default = BuildWin

//var Default = build.BuildAll
