package gui

import (
	"fmt"
	"image"
	"image/color"

	"fyne.io/fyne/v2/canvas"
)

const (
	ScreenWidth  = 240
	ScreenHeight = 160
)

type GameCanvas struct {
	image  *canvas.Image
	pixels []uint8
	scale  int
}

func NewGameCanvas(scale int) *GameCanvas {
	if scale < 1 {
		scale = 1
	}
	if scale > 4 {
		scale = 4
	}

	fmt.Printf("[Canvas] Creating GameCanvas with scale %d, resolution %dx%d\n",
		scale, ScreenWidth*scale, ScreenHeight*scale)

	gc := &GameCanvas{
		scale:  scale,
		pixels: make([]uint8, ScreenWidth*ScreenHeight*4),
	}

	// 创建 Fyne 图像
	gc.image = canvas.NewImageFromImage(gc.createImage())
	gc.image.FillMode = canvas.ImageFillOriginal
	gc.image.ScaleMode = canvas.ImageScalePixels

	fmt.Printf("[Canvas] GameCanvas created successfully\n")

	return gc
}

func (gc *GameCanvas) createImage() image.Image {
	return image.NewRGBA(image.Rect(0, 0, ScreenWidth, ScreenHeight))
}

func (gc *GameCanvas) Image() *canvas.Image {
	return gc.image
}

func (gc *GameCanvas) SetScale(scale int) {
	if scale >= 1 && scale <= 4 {
		gc.scale = scale
	}
}

func (gc *GameCanvas) GetScale() int {
	return gc.scale
}

func (gc *GameCanvas) UpdateFrame(frameBuffer []uint16) {
	if len(frameBuffer) != ScreenWidth*ScreenHeight {
		fmt.Printf("[Canvas] ERROR: Invalid frame buffer size: %d (expected %d)\n",
			len(frameBuffer), ScreenWidth*ScreenHeight)
		return
	}

	// 检查帧缓冲是否全为0（黑屏）
	allZero := true
	for i := 0; i < 100 && i < len(frameBuffer); i++ {
		if frameBuffer[i] != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		fmt.Printf("[Canvas] WARNING: Frame buffer appears to be all zeros (black screen)\n")
	}

	// 将 RGB565 转换为 RGBA
	for i, color16 := range frameBuffer {
		r := uint8((color16 & 0x1F) << 3)
		g := uint8(((color16 >> 5) & 0x1F) << 3)
		b := uint8(((color16 >> 10) & 0x1F) << 3)

		idx := i * 4
		gc.pixels[idx] = r
		gc.pixels[idx+1] = g
		gc.pixels[idx+2] = b
		gc.pixels[idx+3] = 255
	}

	// 创建新图像
	img := image.NewRGBA(image.Rect(0, 0, ScreenWidth, ScreenHeight))
	copy(img.Pix, gc.pixels)

	// 如果缩放 > 1，进行缩放
	if gc.scale > 1 {
		img = gc.scaleImage(img, gc.scale)
	}

	gc.image.Image = img
	gc.image.Refresh()
}

func (gc *GameCanvas) scaleImage(src *image.RGBA, scale int) *image.RGBA {
	width := src.Bounds().Dx() * scale
	height := src.Bounds().Dy() * scale

	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < ScreenHeight; y++ {
		for x := 0; x < ScreenWidth; x++ {
			srcIdx := (y*ScreenWidth + x) * 4
			r := gc.pixels[srcIdx]
			g := gc.pixels[srcIdx+1]
			b := gc.pixels[srcIdx+2]
			a := gc.pixels[srcIdx+3]

			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					dstX := x*scale + dx
					dstY := y*scale + dy
					dstIdx := (dstY*width + dstX) * 4
					dst.Pix[dstIdx] = r
					dst.Pix[dstIdx+1] = g
					dst.Pix[dstIdx+2] = b
					dst.Pix[dstIdx+3] = a
				}
			}
		}
	}

	return dst
}

func (gc *GameCanvas) Clear() {
	for i := range gc.pixels {
		gc.pixels[i] = 0
	}

	img := image.NewRGBA(image.Rect(0, 0, ScreenWidth, ScreenHeight))
	gc.image.Image = img
	gc.image.Refresh()
}

func rgb565ToRGBA(color16 uint16) color.RGBA {
	r := uint8((color16 & 0x1F) << 3)
	g := uint8(((color16 >> 5) & 0x1F) << 3)
	b := uint8(((color16 >> 10) & 0x1F) << 3)
	return color.RGBA{r, g, b, 255}
}
