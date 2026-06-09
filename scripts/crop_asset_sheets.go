package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

var sceneNames = []string{
	"birth-chamber",
	"east-palace-study",
	"winter-hunt",
	"flood-levee",
	"succession-hall",
	"throne-court",
	"granary-relief",
	"tax-office",
	"frontier-fortress",
	"envoy-pass",
	"reform-archive",
	"secret-tribunal",
	"banquet-hall",
	"jiangnan-canal",
	"northern-battlefield",
	"desert-market",
	"imperial-garden",
	"rain-corridor",
	"ancestral-temple",
	"ministry-office",
	"dockyard-fleet",
	"drill-ground",
	"rebel-village",
	"silk-market",
	"mountain-monastery",
	"exam-hall",
	"map-room",
	"palace-dawn",
	"diplomatic-tent",
	"festival-night",
}

var portraitNames = []string{
	"infant-prince",
	"teen-prince",
	"young-emperor",
	"elder-emperor",
	"stern-tutor",
	"frontier-general",
	"finance-minister",
	"grand-princess",
	"noble-consort",
	"young-empress",
	"queen-dowager",
	"palace-maid",
	"eunuch-spymaster",
	"scholar-official",
	"reformist-official",
	"corrupt-magistrate",
	"merchant-leader",
	"foreign-envoy",
	"nomad-khan",
	"monk-strategist",
	"female-diplomat",
	"guard-captain",
	"rebel-leader",
	"river-engineer",
	"imperial-physician",
	"astrologer",
	"poet",
	"court-painter",
	"farmer-representative",
	"masked-assassin",
}

func main() {
	if err := cropSheet("web/assets/sheets/scenes-sheet.png", "web/assets/scenes", "scene", 5, 6, 7, sceneNames); err != nil {
		panic(err)
	}
	if err := cropSheet("web/assets/sheets/portraits-sheet.png", "web/assets/portraits", "portrait", 6, 5, 30, portraitNames); err != nil {
		panic(err)
	}
}

func cropSheet(source, outDir, prefix string, cols, rows, insetPixels int, names []string) error {
	if len(names) != cols*rows {
		return fmt.Errorf("%s needs %d names, got %d", prefix, cols*rows, len(names))
	}

	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	for row := range rows {
		for col := range cols {
			idx := row*cols + col
			x0 := bounds.Min.X + col*width/cols
			x1 := bounds.Min.X + (col+1)*width/cols
			y0 := bounds.Min.Y + row*height/rows
			y1 := bounds.Min.Y + (row+1)*height/rows
			if col == cols-1 {
				x1 = bounds.Max.X
			}
			if row == rows-1 {
				y1 = bounds.Max.Y
			}
			rect := inset(image.Rect(x0, y0, x1, y1), insetPixels)
			name := fmt.Sprintf("%s-%02d-%s.png", prefix, idx+1, names[idx])
			if err := writeCrop(img, rect, filepath.Join(outDir, name)); err != nil {
				return err
			}
		}
	}
	return nil
}

func inset(rect image.Rectangle, pixels int) image.Rectangle {
	if rect.Dx() <= pixels*2 || rect.Dy() <= pixels*2 {
		return rect
	}
	return image.Rect(rect.Min.X+pixels, rect.Min.Y+pixels, rect.Max.X-pixels, rect.Max.Y-pixels)
}

func writeCrop(img image.Image, rect image.Rectangle, out string) error {
	dst := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	draw.Draw(dst, dst.Bounds(), img, rect.Min, draw.Src)
	dst = trimLightBorder(dst)

	file, err := os.Create(out)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, dst)
}

func trimLightBorder(src *image.RGBA) *image.RGBA {
	bounds := src.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if !isNearlyWhite(src.RGBAAt(x, y)) {
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if maxX <= minX || maxY <= minY {
		return src
	}
	trimmedRect := image.Rect(max(0, minX-1), max(0, minY-1), min(bounds.Max.X, maxX+2), min(bounds.Max.Y, maxY+2))
	dst := image.NewRGBA(image.Rect(0, 0, trimmedRect.Dx(), trimmedRect.Dy()))
	draw.Draw(dst, dst.Bounds(), src, trimmedRect.Min, draw.Src)
	return dst
}

func isNearlyWhite(c color.RGBA) bool {
	return c.R > 245 && c.G > 245 && c.B > 245
}
