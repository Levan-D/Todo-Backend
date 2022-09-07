package config

import (
	"github.com/gobuffalo/packr/v2"
)

var BOX *packr.Box

func GetBox() *packr.Box {
	return BOX
}

func SetBox(box *packr.Box) {
	BOX = box
}
