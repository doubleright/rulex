// SPDX-FileCopyrightText: 2020 Kent Gibson <warthog618@gmail.com>
//
// SPDX-License-Identifier: MIT

//go:build linux
// +build linux

// A simple example that toggles an output pin.
package test

import (
	"fmt"
	"os"

	"github.com/warthog618/config"
	"github.com/warthog618/config/blob"
	"github.com/warthog618/config/blob/decoder/json"
	"github.com/warthog618/config/dict"
	"github.com/warthog618/config/env"
	"github.com/warthog618/config/pflag"
	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"
	"github.com/warthog618/gpiod/spi/adc0832"
)

// This example reads both channels from an ADC0832 connected to the RPI by four
// data lines - CSZ, CLK, DI, and DO. The default pin assignments are defined in
// loadConfig, but can be altered via configuration (env, flag or config file).
// All pins other than DO are outputs so do not run this example on a board
// where those pins serve other purposes.

func Test_GPIO_IIC(t *testing.T) {
	cfg := loadConfig()
	tclk := cfg.MustGet("tclk").Duration()
	tset := cfg.MustGet("tset").Duration()
	if tset < tclk {
		tset = 0
	} else {
		tset -= tclk
	}
	chip := cfg.MustGet("gpiochip").String()
	c, err := gpiod.NewChip(chip, gpiod.WithConsumer("adc0832"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "adc0832: %s\n", err)
		os.Exit(1)
	}
	a, err := adc0832.New(
		c,
		cfg.MustGet("clk").Int(),
		cfg.MustGet("csz").Int(),
		cfg.MustGet("di").Int(),
		cfg.MustGet("do").Int(),
		adc0832.WithTclk(tclk),
		adc0832.WithTset(tset),
	)
	c.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "adc0832: %s\n", err)
		os.Exit(1)
	}
	defer a.Close()
	ch0, err := a.Read(0)
	if err != nil {
		fmt.Printf("read error ch0: %s\n", err)
	}
	ch1, err := a.Read(1)
	if err != nil {
		fmt.Printf("read error ch1: %s\n", err)
	}
	fmt.Printf("ch0=0x%02x, ch1=0x%02x\n", ch0, ch1)
}

func loadConfig() *config.Config {
	defaultConfig := map[string]interface{}{
		"gpiochip": "gpiochip0",
		"tclk":     "2500ns",
		"tset":     "2500ns", // should be at least tclk - enforced in main
		"csz":      rpi.J8p29,
		"clk":      rpi.J8p31,
		"do":       rpi.J8p33,
		"di":       rpi.J8p35,
	}
	def := dict.New(dict.WithMap(defaultConfig))
	flags := []pflag.Flag{
		{Short: 'c', Name: "config-file"},
	}
	cfg := config.New(
		pflag.New(pflag.WithFlags(flags)),
		env.New(env.WithEnvPrefix("ADC0832_")),
		config.WithDefault(def))
	cfg.Append(
		blob.NewConfigFile(cfg, "config.file", "adc0832.json", json.NewDecoder()))
	cfg = cfg.GetConfig("", config.WithMust())
	return cfg
}

// This example drives GPIO 22, which is pin J8-15 on a Raspberry Pi.
// The pin is toggled high and low at 1Hz with a 50% duty cycle.
// Do not run this on a device which has this pin externally driven.
func Test_GPIO_BLINKER(t *testing.T) {
	offset := rpi.J8p15
	v := 0
	l, err := gpiod.RequestLine("gpiochip0", offset, gpiod.AsOutput(v))
	if err != nil {
		panic(err)
	}
	// revert line to input on the way out.
	defer func() {
		l.Reconfigure(gpiod.AsInput)
		l.Close()
	}()
	values := map[int]string{0: "inactive", 1: "active"}
	fmt.Printf("Set pin %d %s\n", offset, values[v])

	// capture exit signals to ensure pin is reverted to input on exit.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	for {
		select {
		case <-time.After(2 * time.Second):
			v ^= 1
			l.SetValue(v)
			fmt.Printf("Set pin %d %s\n", offset, values[v])
		case <-quit:
			return
		}
	}
}
