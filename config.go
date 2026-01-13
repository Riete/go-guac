package guac

import (
	"strconv"
)

type Config struct {
	Protocol     string
	Params       map[string]string
	ScreenWidth  int
	ScreenHeight int
	ScreenDpi    int
}

func (c *Config) SetScreen() {
	c.Params["width"] = strconv.Itoa(c.ScreenWidth)
	c.Params["height"] = strconv.Itoa(c.ScreenHeight)
	c.Params["dpi"] = strconv.Itoa(c.ScreenDpi)
}

func NewDefaultConfig() *Config {
	return &Config{
		Protocol:     "rdp",
		Params:       make(map[string]string),
		ScreenWidth:  1920,
		ScreenHeight: 1080,
		ScreenDpi:    96,
	}
}
