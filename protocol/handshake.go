package protocol

import (
	"strconv"
)

const defaultScreenWidth = 1920
const defaultScreenHeight = 1080
const defaultScreenDpi = 96
const defaultProtocol = "rdp"

type HandshakeOption func(*HandshakeConfig)

func WithProtocol(protocol string) HandshakeOption {
	return func(c *HandshakeConfig) {
		c.protocol = protocol
	}
}

func WithDomain(domain string) HandshakeOption {
	return func(c *HandshakeConfig) {
		c.connectArgs["domain"] = domain
	}
}

func WithAuth(username, password string) HandshakeOption {
	return func(c *HandshakeConfig) {
		c.connectArgs["username"] = username
		c.connectArgs["password"] = password
	}
}

func WithHostPort(host, port string) HandshakeOption {
	return func(c *HandshakeConfig) {
		c.connectArgs["hostname"] = host
		c.connectArgs["port"] = port
	}
}

func WithScreen(width, height, dpi int) HandshakeOption {
	return func(c *HandshakeConfig) {
		c.width = width
		c.height = height
		c.dpi = dpi
	}
}

func WithAudioCodecs(codecs []string) HandshakeOption {
	return func(c *HandshakeConfig) {
		c.audioCodecs = codecs
	}
}

func WithVideoCodecs(codecs []string) HandshakeOption {
	return func(c *HandshakeConfig) {
		c.videoCodecs = codecs
	}
}

func WithImageFormats(formats []string) HandshakeOption {
	return func(c *HandshakeConfig) {
		c.imageFormats = formats
	}
}

func WithIgnoreCert() HandshakeOption {
	return func(c *HandshakeConfig) {
		c.connectArgs["ignore-cert"] = "true"
	}
}

func WithSecurity(security string) HandshakeOption {
	return func(c *HandshakeConfig) {
		c.connectArgs["security"] = security
	}
}

func WithNLASecurity() HandshakeOption {
	return func(c *HandshakeConfig) {
		c.connectArgs["security"] = "nla"
	}
}

func WithReadOnly() HandshakeOption {
	return func(c *HandshakeConfig) {
		c.connectArgs["read-only"] = "true"
	}
}

type HandshakeConfig struct {
	protocol     string
	connectArgs  map[string]string
	width        int
	height       int
	dpi          int
	audioCodecs  []string
	videoCodecs  []string
	imageFormats []string
}

func (h *HandshakeConfig) setScreen() {
	h.connectArgs["width"] = strconv.Itoa(h.width)
	h.connectArgs["height"] = strconv.Itoa(h.height)
	h.connectArgs["dpi"] = strconv.Itoa(h.dpi)
}

func (h *HandshakeConfig) SelectInstruction() Instruction {
	return NewInstruction("select", h.protocol)
}

func (h *HandshakeConfig) SizeInstruction() Instruction {
	return NewInstruction("size", strconv.Itoa(h.width), strconv.Itoa(h.height), strconv.Itoa(h.dpi))
}

func (h *HandshakeConfig) ConnectInstruction(args []Element) Instruction {
	var argsValues []string
	for _, arg := range args {
		argsValues = append(argsValues, h.connectArgs[arg.Value()])
	}
	return NewInstruction("connect", argsValues...)
}

func (h *HandshakeConfig) AudioInstruction() Instruction {
	return NewInstruction("audio", h.audioCodecs...)
}

func (h *HandshakeConfig) VideoInstruction() Instruction {
	return NewInstruction("video", h.videoCodecs...)
}

func (h *HandshakeConfig) ImageInstruction() Instruction {
	return NewInstruction("image", h.imageFormats...)
}

func NewHandshakeConfig(connectArgs map[string]string, opts ...HandshakeOption) *HandshakeConfig {
	if connectArgs == nil {
		connectArgs = make(map[string]string)
	}
	config := &HandshakeConfig{
		protocol:    defaultProtocol,
		connectArgs: connectArgs,
		width:       defaultScreenWidth,
		height:      defaultScreenHeight,
		dpi:         defaultScreenDpi,
	}
	for _, opt := range opts {
		opt(config)
	}
	config.setScreen()
	return config
}
