package buildpack

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

type Version2Pick struct {
	Label    string
	Version  string
	DPI      float64
	font     *truetype.Font
	FontSize float64
	Margin   int
}

var (
	labelBg    = color.RGBA{15, 186, 236, 0xff}
	verBg      = color.RGBA{225, 226, 227, 0xff}
	labelColor = image.White
	verColor   = image.NewUniform(color.RGBA{39, 39, 39, 0xff})
)

func DefaultVer2Pick(label, version string) Version2Pick {
	return Version2Pick{
		Label:    label,
		Version:  version,
		DPI:      150,
		FontSize: 14,
		Margin:   10,
	}
}

func (v *Version2Pick) calculateSizeOfText(text string) (int, int) {
	c := freetype.NewContext()
	c.SetDPI(v.DPI)
	c.SetFont(v.font)
	c.SetFontSize(v.FontSize)
	c.SetSrc(image.White)
	pt := freetype.Pt(0, 0)
	rpt, err := c.DrawString(text, pt)
	if err != nil {
		log.Println(err)
		return 0, 0
	}
	return rpt.X.Floor(), c.PointToFixed(v.FontSize).Floor()
}

func (v *Version2Pick) Generate(dir string) error {
	fontBytes, err := hex.DecodeString(fontHex)
	if err != nil {
		return err
	}
	v.font, err = freetype.ParseFont(fontBytes)
	if err != nil {
		return err
	}

	labelW, labelH := v.calculateSizeOfText(v.Label)
	verW, _ := v.calculateSizeOfText(v.Version)

	width := labelW + verW + v.Margin*4
	height := labelH + v.Margin
	separated := labelW + v.Margin*2

	background := image.NewRGBA(image.Rect(0, 0, width, height))

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if x <= separated {
				background.SetRGBA(x, y, labelBg)
			} else {
				background.SetRGBA(x, y, verBg)
			}
		}
	}
	{
		c := freetype.NewContext()
		c.SetDPI(v.DPI)
		c.SetFont(v.font)
		c.SetFontSize(v.FontSize)
		c.SetClip(background.Bounds())
		c.SetDst(background)
		c.SetSrc(labelColor)
		pt := freetype.Pt(v.Margin, labelH)
		_, err := c.DrawString(v.Label, pt)
		if err != nil {
			return err
		}
	}

	{
		c := freetype.NewContext()
		c.SetDPI(v.DPI)
		c.SetFont(v.font)
		c.SetFontSize(v.FontSize)
		c.SetClip(background.Bounds())
		c.SetDst(background)
		c.SetSrc(verColor)
		pt := freetype.Pt(separated+v.Margin, labelH)
		_, err = c.DrawString(v.Version, pt)
		if err != nil {
			return err
		}
	}

	// Save that RGBA image to disk.
	outFile, err := os.Create(filepath.Join(dir, fmt.Sprintf("VERSION_%s", v.Label)))
	if err != nil {
		return err
	}
	defer func() {
		_ = outFile.Close()
	}()
	b := bufio.NewWriter(outFile)
	err = png.Encode(b, background)
	if err != nil {
		return err
	}
	err = b.Flush()
	if err != nil {
		return err
	}
	return nil
}
