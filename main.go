package main

import (
	"math"
	"math/rand"
	"strconv"
	"syscall/js"
)

const pointCount = 60

const black = "#000000"
const white = "#FFFFFF"
const gray = "#444444"
const font = "8px sans-serif"
const framePrefix = "frame="

//var frameSB strings.Builder
var frameNo = 0

var animator js.Func

type Paper struct {
	ctx    js.Value
	Width  int
	Height int
}
type Point struct {
	x   float32
	y   float32
	dxv float32
	dyv float32
}

var points [pointCount]Point

func createCanvasWithID(id string) *Paper {
	doc := js.Global().Get("document")
	scale := js.Global().Get("devicePixelRatio").Float()
	app := doc.Call("getElementById", "app")
	app.Set("style", "float:left; width:100%; height:100%;")
	cw := app.Get("clientWidth").Int()
	ch := app.Get("clientHeight").Int()
	actualWidth := int(float64(cw) * scale)
	actualHeight := int(float64(ch) * scale)

	// create canvas
	canvas := doc.Call("createElement", "canvas")
	canvas.Set("id", id)
	canvas.Set("width", actualWidth)
	canvas.Set("height", actualHeight)

	style := canvas.Get("style")

	style.Set("width", strconv.Itoa(cw)+"px")
	style.Set("height", strconv.Itoa(ch)+"px")
	style.Set("background-color", "#777777")

	ctx := canvas.Call("getContext", "2d", `{ alpha: false }`)
	ctx.Call("scale", scale, scale)
	app.Call("appendChild", canvas)
	return &Paper{ctx, cw, ch}
}

func main() {
	//rand seed (from js so we dont need time pkg)
	d := js.Global().Get("Date")
	t := d.Call("now").Float() //64b
	rand.Seed(int64(t))

	//create the canvas
	paper := createCanvasWithID("canvas")
	ctx := paper.ctx
	tx := 0 //TODO from tile based demo code - some have been removed
	ty := 0
	tw := paper.Width
	th := paper.Height
	var gridSizeX = tw / 8
	var gridSizeY = th / 4
	var maxStarDistance = float32(gridSizeX) * .8
	var cols = (tw / gridSizeX)
	var rows = (th / gridSizeY)
	var grid = make([][][]Point, cols)
	for i := range grid {
		grid[i] = make([][]Point, rows)
		for j := range grid[i] {
			grid[i][j] = make([]Point, 0)
		}
	}

	//create stars
	for i := 0; i < pointCount; i++ {
		px := (rand.Float32() * float32(paper.Width))
		py := (rand.Float32() * float32(paper.Height))
		points[i] = Point{px, py, (rand.Float32() - .5), (rand.Float32() - .5)}
	}

	//animator
	animator = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		frameNo++
		ctx.Call("save")

		//build star grids
		for x := 0; x < cols; x++ {
			for y := 0; y < rows; y++ {
				grid[x][y] = grid[x][y][:0]
			}
		}
		for _, p := range points {
			gx := int((p.x - float32(tx)) / float32(gridSizeX))
			gy := int((p.y - float32(ty)) / float32(gridSizeY))
			if gx >= cols || gy >= rows || gx < 0 || gy < 0 {
				// 	//println("out of bounds:  gx=$gx ($cols) gy=$gy ($rows)")
				continue
			}
			grid[gx][gy] = append(grid[gx][gy], p)
		}

		//clear canvas
		ctx.Set("fillStyle", black)
		ctx.Set("strokeStyle", black)
		ctx.Set("lineWidth", 1)
		ctx.Call("strokeRect", tx, ty, tw, th)
		ctx.Call("fillRect", tx, ty, tw, th)

		//frames
		// ctx.Set("lineWidth", .5)
		// ctx.Set("fillStyle", white)
		// ctx.Set("font", font)
		// frameSB.Reset()
		// frameSB.WriteString(framePrefix)
		// frameSB.WriteString(strconv.Itoa(frames))
		// ctx.Call("fillText", frameSB.String(), tx+20, ty+20)

		//draw lines
		ctx.Set("lineWidth", 1)
		ctx.Set("strokeStyle", gray)
		for x := 0; x < cols; x++ {
			for y := 0; y < rows; y++ {
				//TODO dont draw duplicates
				for _, p1 := range grid[x][y] {
					for _, p2 := range grid[x][y] {
						d := float32(math.Sqrt(float64(((p1.x - p2.x) * (p1.x - p2.x)) + ((p1.y - p2.y) * (p1.y - p2.y)))))
						if d < maxStarDistance {
							ctx.Call("beginPath")
							ctx.Call("moveTo", p1.x, p1.y)
							ctx.Call("lineTo", p2.x, p2.y)
							ctx.Call("stroke")
						}
					}
				}
			}
		}

		//draw stars
		ctx.Set("fillStyle", white)
		for i := 0; i < pointCount; i++ {
			ctx.Call("beginPath")
			ctx.Call("arc", points[i].x, points[i].y, 1.0, 0.0, 2*3.14159265358979323846, true)
			ctx.Call("fill")
		}

		//move stars
		for i := 0; i < pointCount; i++ {
			points[i].x = float32(math.Mod((float64(points[i].x) + float64(tw) + float64(points[i].dxv)), float64(tw)))
			points[i].y = float32(math.Mod((float64(points[i].y) + float64(th) + float64(points[i].dyv)), float64(th)))
		}
		ctx.Call("restore")
		js.Global().Call("requestAnimationFrame", animator)
		return nil
	})
	//defer animationCallback.Release()
	js.Global().Call("requestAnimationFrame", animator)
}
