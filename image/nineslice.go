package image

import (
	"image"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// A NineSlice is an image that can be drawn with any width and height. It is basically a 3x3 grid of image tiles:
// The corner tiles are drawn as-is, while the center columns and rows of tiles will be stretched to fit the desired
// width and height.
type NineSlice struct {
	image       *ebiten.Image
	widths      [3]int
	heights     [3]int
	transparent bool

	init  sync.Once
	tiles [9]*ebiten.Image
}

type borders struct {
	borderTop    int
	borderRight  int
	borderBottom int
	borderLeft   int
	borderColor  color.Color
}

// This function returns a new borders struct instance for use with the NewAdvancedNineSlice* functions.
func NewBorder(borderTop int, borderRight int, borderBottom int, borderLeft int, borderColor color.Color) borders {
	return borders{max(borderTop, 0), max(borderRight, 0), max(borderBottom, 0), max(borderLeft, 0), borderColor}
}

// A DrawImageOptionsFunc is responsible for setting DrawImageOptions when drawing an image.
// This is usually used to translate the image.
type DrawImageOptionsFunc func(opts *ebiten.DrawImageOptions)

var colorImages map[color.Color]*ebiten.Image = map[color.Color]*ebiten.Image{}

var colorNineSlices map[color.Color]*NineSlice = map[color.Color]*NineSlice{}

// NewNineSlice constructs a new NineSlice from i, having columns widths w and row heights h.
func NewNineSlice(i *ebiten.Image, w [3]int, h [3]int) *NineSlice {
	return &NineSlice{
		image:   i,
		widths:  w,
		heights: h,
	}
}

func NewFixedNineSlice(i *ebiten.Image) *NineSlice {
	return &NineSlice{
		image:   i,
		widths:  [3]int{i.Bounds().Dx(), 0, 0},
		heights: [3]int{i.Bounds().Dy(), 0, 0},
	}
}

// NewNineSliceSimple constructs a new NineSlice from image. borderWidthHeight specifies the width of the
// left and right column and the height of the top and bottom row. centerWidthHeight specifies the width
// of the center column and row.
func NewNineSliceSimple(image *ebiten.Image, borderWidthHeight int, centerWidthHeight int) *NineSlice {
	return &NineSlice{
		image:   image,
		widths:  [3]int{borderWidthHeight, centerWidthHeight, borderWidthHeight},
		heights: [3]int{borderWidthHeight, centerWidthHeight, borderWidthHeight},
	}
}

// NewNineSliceSimple constructs a new NineSlice from image. borderWidthHeight specifies the width of the
// left and right column and the height of the top and bottom row. The center width and height is computed as
// the width of image minus 2*borderWidthHeight.
func NewNineSliceBorder(image *ebiten.Image, borderWidthHeight int) *NineSlice {
	return NewNineSliceSimple(image, borderWidthHeight, image.Bounds().Dx()-borderWidthHeight*2)
}

// NewNineSliceColor constructs a new NineSlice that when drawn fills with color c.
func NewNineSliceColor(c color.Color) *NineSlice {
	if c == nil {
		return &NineSlice{transparent: true}
	}

	if n, ok := colorNineSlices[c]; ok {
		return n
	}

	var n *NineSlice
	if _, _, _, a := c.RGBA(); a == 0 {
		n = &NineSlice{
			transparent: true,
		}
	} else {
		n = &NineSlice{
			image:   NewImageColor(c),
			widths:  [3]int{0, 1, 0},
			heights: [3]int{0, 1, 0},
		}
	}
	colorNineSlices[c] = n
	return n
}

// NewBorderedNineSliceColor constructs a new NineSlice that when drawn fills with color c and has a border with
// the specified color and width.
func NewBorderedNineSliceColor(bodyColor color.Color, borderColor color.Color, borderWidth int) *NineSlice {
	i := ebiten.NewImage(borderWidth*2+1, borderWidth*2+1)
	i.Fill(borderColor)
	i.Set(borderWidth, borderWidth, bodyColor)

	return &NineSlice{
		image:   i,
		widths:  [3]int{borderWidth, 1, borderWidth},
		heights: [3]int{borderWidth, 1, borderWidth},
	}
}

// NewBorderedNineSliceColor constructs a new NineSlice that when drawn fills with the provided image and has a border with
// the specified color and width.
func NewBorderedNineSliceImage(img *ebiten.Image, borderColor color.Color, borderWidth int) *NineSlice {
	i := ebiten.NewImage(borderWidth*2+img.Bounds().Dx(), borderWidth*2+img.Bounds().Dy())
	// Set the border color.
	i.Fill(borderColor)
	// Draw the image in the middle.
	geo := ebiten.GeoM{}
	geo.Translate(float64(borderWidth), float64(borderWidth))
	i.DrawImage(img, &ebiten.DrawImageOptions{
		GeoM: geo,
	})

	return &NineSlice{
		image:   i,
		widths:  [3]int{borderWidth, img.Bounds().Dx(), borderWidth},
		heights: [3]int{borderWidth, img.Bounds().Dy(), borderWidth},
	}
}

// NewAdvancedNineSliceColor constructs a new NineSlice that when drawn fills with color c and
// has a border defined by the borders struct.
func NewAdvancedNineSliceColor(bodyColor color.Color, border borders) *NineSlice {
	i := ebiten.NewImage(border.borderLeft+border.borderRight+1, border.borderTop+border.borderBottom*2+1)
	i.Fill(border.borderColor)
	i.Set(border.borderLeft, border.borderTop, bodyColor)

	return &NineSlice{
		image:   i,
		widths:  [3]int{border.borderLeft, 1, border.borderRight},
		heights: [3]int{border.borderTop, 1, border.borderBottom},
	}
}

// NewAdvancedNineSliceImage constructs a new NineSlice that when drawn fills with the provided image and
// has a border defined by the borders struct.
func NewAdvancedNineSliceImage(img *ebiten.Image, border borders) *NineSlice {
	i := ebiten.NewImage(border.borderLeft+border.borderRight+img.Bounds().Dx(), border.borderTop+border.borderBottom+img.Bounds().Dy())
	// Set the border color.
	i.Fill(border.borderColor)
	// Draw the image in the middle.
	geo := ebiten.GeoM{}
	geo.Translate(float64(border.borderLeft), float64(border.borderTop))
	i.DrawImage(img, &ebiten.DrawImageOptions{
		GeoM: geo,
	})

	return &NineSlice{
		image:   i,
		widths:  [3]int{border.borderLeft, img.Bounds().Dx(), border.borderRight},
		heights: [3]int{border.borderTop, img.Bounds().Dy(), border.borderBottom},
	}
}

// NewImageColor constructs a new Image that when drawn fills with color c.
func NewImageColor(c color.Color) *ebiten.Image {
	if i, ok := colorImages[c]; ok {
		return i
	}

	i := ebiten.NewImage(1, 1)
	i.Fill(c)
	colorImages[c] = i
	return i
}

// Draw draws n onto screen, with the size specified by width and height. If optsFunc is not nil, it is used to set
// DrawImageOptions for each tile drawn.
func (n *NineSlice) Draw(screen *ebiten.Image, width int, height int, optsFunc DrawImageOptionsFunc) {
	if n.transparent {
		return
	}

	n.drawTiles(screen, width, height, optsFunc)
}

func (n *NineSlice) drawTiles(screen *ebiten.Image, width int, height int, optsFunc DrawImageOptionsFunc) {
	n.init.Do(n.createTiles)

	sy := 0
	ty := 0
	for r, sh := range n.heights {
		sx := 0
		tx := 0

		var th int
		if r == 1 {
			th = height - n.heights[0] - n.heights[2]
		} else {
			th = sh
		}

		for c, sw := range n.widths {
			var tw int
			if c == 1 {
				tw = width - n.widths[0] - n.widths[2]
			} else {
				tw = sw
			}

			n.drawTile(screen, n.tiles[r*3+c], tx, ty, sw, sh, tw, th, optsFunc)

			sx += sw
			tx += tw
		}

		sy += sh
		ty += th
	}
}

func (n *NineSlice) drawTile(screen *ebiten.Image, tile *ebiten.Image, tx int, ty int, sw int, sh int, tw int, th int, optsFunc DrawImageOptionsFunc) {
	if sw <= 0 || sh <= 0 || tw <= 0 || th <= 0 {
		return
	}

	opts := ebiten.DrawImageOptions{
		Filter: ebiten.FilterNearest,
	}

	if tw != sw || th != sh {
		opts.GeoM.Scale(float64(tw)/float64(sw), float64(th)/float64(sh))
	}

	opts.GeoM.Translate(float64(tx), float64(ty))

	if optsFunc != nil {
		optsFunc(&opts)
	}

	screen.DrawImage(tile, &opts)
}

func (n *NineSlice) createTiles() {
	defer func() {
		n.image = nil
	}()

	n.tiles = [9]*ebiten.Image{}

	if n.centerOnly() {
		n.tiles[1*3+1] = n.image
		return
	}

	minSize := n.image.Bounds().Min

	sy := minSize.Y
	for r, sh := range n.heights {
		sx := minSize.X
		for c, sw := range n.widths {
			if sh > 0 && sw > 0 {
				rect := image.Rect(0, 0, sw, sh)
				rect = rect.Add(image.Point{sx, sy})
				if tile, ok := n.image.SubImage(rect).(*ebiten.Image); ok {
					n.tiles[r*3+c] = tile
				}
			}
			sx += sw
		}
		sy += sh
	}
}

func (n *NineSlice) centerOnly() bool {
	if n.widths[0] > 0 || n.widths[2] > 0 || n.heights[0] > 0 || n.heights[2] > 0 {
		return false
	}

	w, h := n.image.Bounds().Dx(), n.image.Bounds().Dy()
	return n.widths[1] == w && n.heights[1] == h
}

// MinSize returns the minimum width and height to draw n correctly. If n is drawn with a smaller size,
// the corner or edge tiles will overlap.
func (n *NineSlice) MinSize() (int, int) {
	if n.transparent {
		return 0, 0
	}

	return n.widths[0] + n.widths[2], n.heights[0] + n.heights[2]
}
