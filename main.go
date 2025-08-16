package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"log"

	"screen-rule/assets/fonts"

	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 400
	screenHeight = 180
)

var (
	brushImage                  *ebiten.Image
	monitorWidth, monitorHeight = ebiten.Monitor().Size()
	shsFaceSource               *text.GoTextFaceSource
)

type pos struct {
	x int
	y int
}

type Game struct {
	cursor           pos
	dragging         bool
	dragStartWindowX int
	dragStartWindowY int
	dragStartCursorX int
	dragStartCursorY int

	cursorToWindowX float64
	cursorToWindowY float64

	canvasImage *ebiten.Image

	debugui   debugui.DebugUI
	debugRect image.Rectangle
	count     int
}

func NewGame() *Game {
	g := &Game{
		canvasImage: ebiten.NewImage(screenWidth, screenHeight),
	}
	g.canvasImage.Fill(color.Transparent)
	g.cursor = pos{
		x: -1,
		y: -1,
	}
	return g
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	mx, my := ebiten.CursorPosition()
	isMouseMoved := g.cursor.x != mx || g.cursor.y != my
	g.cursor = pos{
		x: mx,
		y: my,
	}
	isMouseChanged := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) ||
		inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle) ||
		inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonMiddle) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) ||
		inpututil.IsMouseButtonJustReleased(ebiten.MouseButton1) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButton2) ||
		inpututil.IsMouseButtonJustReleased(ebiten.MouseButton3) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButton4)
	if isMouseChanged || isMouseMoved {
		g.canvasImage.Fill(color.Transparent)
		// g.canvasImage.Fill(color.White) // for debug
		g.drawText(g.canvasImage, "你好世界！", 50, 146, 5)
		g.drawText(g.canvasImage, "测试", 56, 90, 80)

		b := g.canvasImage.Bounds()
		var ebitenAlphaImage *image.Alpha = image.NewAlpha(b)
		for j := b.Min.Y; j < b.Max.Y; j++ {
			for i := b.Min.X; i < b.Max.X; i++ {
				ebitenAlphaImage.Set(i, j, g.canvasImage.At(i, j))
			}
		}
		isIn := ebitenAlphaImage.At(mx-b.Min.X, my-b.Min.Y).(color.Alpha).A > 0 || g.isOnDebugUI(mx, my)
		ebiten.SetWindowMousePassthrough(!isIn)
	}

	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.dragging = false
	}
	if !g.dragging && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.dragging = true
		g.dragStartWindowX, g.dragStartWindowY = ebiten.WindowPosition()
		g.dragStartCursorX, g.dragStartCursorY = ebiten.CursorPosition()
	}
	if g.dragging {
		// Move the window only by the delta of the cursor.
		cx, cy := ebiten.CursorPosition()
		dx := int(float64(cx-g.dragStartCursorX) * g.cursorToWindowX)
		dy := int(float64(cy-g.dragStartCursorY) * g.cursorToWindowY)
		wx, wy := ebiten.WindowPosition()
		ebiten.SetWindowPosition(wx+dx, wy+dy)
	}
	return g.updateDebugUI()
}

func (g *Game) updateDebugUI() error {
	if _, err := g.debugui.Update(func(ctx *debugui.Context) error {
		if g.debugRect.Empty() {
			g.debugRect = image.Rect(250, 100, 350, 150)
		}
		ctx.Window("Test", g.debugRect, func(layout debugui.ContainerLayout) {
			ctx.Button("Button").On(func() {
				g.count++
			})
		})
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (g *Game) isOnDebugUI(cx, cy int) bool {
	if g.debugRect.Empty() {
		return false
	}
	return g.debugRect.Min.X <= cx && cx < g.debugRect.Max.X &&
		g.debugRect.Min.Y <= cy && cy < g.debugRect.Max.Y
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.canvasImage, nil)

	op := &colorm.DrawImageOptions{}
	op.GeoM.Translate(float64(g.cursor.x-2), float64(g.cursor.y-4))
	var cm colorm.ColorM
	cm.Scale(1.0, 0.50, 0.125, 1.0)
	colorm.DrawImage(screen, brushImage, cm, op)

	isMP := ebiten.IsWindowMousePassthrough()
	msg := fmt.Sprintf(
		"(%d, %d)\n(%d, %d)\nTPS: %0.2f\nMousePassthrough: %v\nHello, World!\nCount: %d",
		monitorWidth, monitorHeight, g.cursor.x, g.cursor.y, ebiten.ActualTPS(), isMP, g.count)
	ebitenutil.DebugPrintAt(screen, msg, 10, 10)

	vector.StrokeRect(g.canvasImage, 2, 2, screenWidth-4, screenHeight-4, 2, color.RGBA{R: 0xff, G: 0xaa, B: 0x11, A: 0xff}, true)

	g.debugui.Draw(screen)
}

func (g *Game) drawText(parent *ebiten.Image, content string, fontsize float64, posX, posY float64) {
	tf := &text.GoTextFace{
		Source: shsFaceSource,
		Size:   fontsize,
	}
	tf.SetVariation(text.MustParseTag("wght"), float32(text.WeightExtraBold)) // 字重
	tf.SetVariation(text.MustParseTag("wdth"), 100)                           // 字宽
	// tf.SetVariation(text.MustParseTag("ital"), 1)                        // 斜体
	// tf.SetVariation(text.MustParseTag("slnt"), 1)                        // 倾斜
	// tf.SetVariation(text.MustParseTag("opsz"), 24)                       // 字体大小

	mw, mh := text.Measure(content, tf, 0)
	tima := image.NewAlpha(image.Rectangle{
		Min: image.Point{
			X: int(0),
			Y: int(0),
		},
		Max: image.Point{
			X: int(mw + 10),
			Y: int(mh + 10),
		},
	})
	timg := ebiten.NewImageFromImage(tima)
	timg.Fill(color.Transparent)
	// timg.Fill(color.White)

	opt := &text.DrawOptions{}
	opt.GeoM.Translate(5, 5)
	opt.ColorScale.ScaleWithColor(color.RGBA{R: 0xff, G: 0xaa, B: 0x11, A: 0xff})
	text.Draw(timg, content, tf, opt)

	opi := &ebiten.DrawImageOptions{}
	opi.GeoM.Translate(posX, posY)
	parent.DrawImage(timg, opi)
	timg.Deallocate()
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.cursorToWindowX = float64(outsideWidth) / float64(screenWidth)
	g.cursorToWindowY = float64(outsideHeight) / float64(screenHeight)
	return screenWidth, screenHeight
}

func init() {
	const (
		a0 = 0x40
		a1 = 0xc0
		a2 = 0xff
	)
	pixels := []uint8{
		a0, a1, a1, a0,
		a1, a2, a2, a1,
		a1, a2, a2, a1,
		a0, a1, a1, a0,
	}
	brushImage = ebiten.NewImageFromImage(&image.Alpha{
		Pix:    pixels,
		Stride: 4,
		Rect:   image.Rect(0, 0, 4, 4),
	})

	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.SourceHanSansSC_VF_ttf))
	if err != nil {
		log.Fatal(err)
	}
	shsFaceSource = s
}

func main() {
	title := "Hello, World!"

	ebiten.SetWindowPosition(0, 0)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetScreenClearedEveryFrame(true)
	ebiten.SetTPS(60)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowMousePassthrough(true)
	ebiten.SetWindowTitle(title)

	game := NewGame()
	options := &ebiten.RunGameOptions{
		InitUnfocused:     true,
		ScreenTransparent: true,
		SkipTaskbar:       true,
		X11ClassName:      title,
		X11InstanceName:   title,
	}
	if err := ebiten.RunGameWithOptions(game, options); err != nil {
		log.Fatal(err)
	}
}
