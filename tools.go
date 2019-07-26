// +build tools

// This file just exists to ensure we download the tools we need for building
// See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

package flux

import (
	_ "github.com/jteeuwen/go-bindata"
	_ "k8s.io/code-generator"
)
