package crop

import (
	"image"
	"image/color"
)

const grayDarknessLimit = 128

func Limited(img image.Image, limit float32) image.Rectangle {
	bounds := img.Bounds()
	maxPixels := float32((bounds.Dx()+bounds.Dy())/2) * limit
	return Bounds(img).Union(bounds.Inset(int(maxPixels)))
}

func Bounds(img image.Image) image.Rectangle {
	left := findBorder(img, image.Pt(1, 0))
	right := findBorder(img, image.Pt(-1, 0))
	top := findBorder(img, image.Pt(0, 1))
	bottom := findBorder(img, image.Pt(0, -1))

	return image.Rect(left.X, top.Y, right.X, bottom.Y)
}

func findBorder(img image.Image, dir image.Point) image.Point {
	bounds := img.Bounds()
	scan := image.Pt(dir.Y, dir.X)
	pt := pointInScanCorner(bounds, dir)

	for !scanLineForNonWhitespace(img, pt, scan) {
		pt = pt.Add(dir)
		if !pt.In(bounds) {
			pt = pointInScanCorner(bounds, dir)
			break
		}
	}

	if dir.X < 0 || dir.Y < 0 {
		return pt.Sub(dir)
	} else {
		return pt
	}
}

func pointInScanCorner(rect image.Rectangle, dir image.Point) image.Point {
	if dir.X < 0 || dir.Y < 0 {
		return rect.Max.Sub(image.Pt(1, 1))
	} else {
		return rect.Min
	}
}

func scanLineForNonWhitespace(img image.Image, pt image.Point, scan image.Point) bool {
	for ; pt.In(img.Bounds()); pt = pt.Add(scan) {
		if gray, ok := color.GrayModel.Convert(img.At(pt.X, pt.Y)).(color.Gray); ok {
			if gray.Y <= grayDarknessLimit {
				return true
			}
		}
	}

	return false
}
