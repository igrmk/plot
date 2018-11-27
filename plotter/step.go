// Copyright ©2018 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plotter

import (
	"image/color"
	"math"

	"github.com/igrmk/plot"
	"github.com/igrmk/plot/vg"
	"github.com/igrmk/plot/vg/draw"
)

// StepKind specifies a form of a connection of two consecutive points.
type StepKind int

const (
	// PreStep connects two points by following lines: vertical, horizontal.
	PreStep StepKind = iota

	// MidStep connects two points by following lines: horizontal, vertical, horizontal.
	// Vertical line is placed in the middle of the interval.
	MidStep

	// PostStep connects two points by following lines: horizontal, vertical.
	PostStep
)

// Step implements the Plotter interface, drawing a stepped line.
type Step struct {
	// XYs is a copy of the points for this line.
	XYs

	// StepStyle is the kind of the step line.
	StepStyle StepKind

	// LineStyle is the style of the line connecting the points.
	// Use nil to not draw the line.
	LineStyle *draw.LineStyle

	// FillColor is the color to fill the area between the x-axis and the plot.
	// Use nil to disable the filling. This is default.
	FillColor color.Color
}

// NewStep returns a Step that uses the default line style and does not draw glyphs.
func NewStep(xys XYer) (*Step, error) {
	data, err := CopyXYs(xys)
	if err != nil {
		return nil, err
	}
	return &Step{
		XYs:       data,
		LineStyle: &DefaultLineStyle,
	}, nil
}

// Plot draws the Step, implementing the plot.Plotter interface.
func (pts *Step) Plot(c draw.Canvas, plt *plot.Plot) {
	trX, trY := plt.Transforms(&c)
	ps := make([]vg.Point, len(pts.XYs))

	for i, p := range pts.XYs {
		ps[i].X = trX(p.X)
		ps[i].Y = trY(p.Y)
	}

	if pts.FillColor != nil && len(ps) > 0 {
		c.SetColor(pts.FillColor)
		minY := trY(plt.Y.Min)
		var pa vg.Path
		pa.Move(vg.Point{X: ps[0].X, Y: minY})
		prev := ps[0]
		if pts.StepStyle != PreStep {
			pa.Line(prev)
		}
		for i, pt := range ps[1:] {
			switch pts.StepStyle {
			case PreStep:
				pa.Line(vg.Point{X: prev.X, Y: pt.Y})
				pa.Line(pt)
			case MidStep:
				pa.Line(vg.Point{X: (prev.X + pt.X) / 2, Y: prev.Y})
				pa.Line(vg.Point{X: (prev.X + pt.X) / 2, Y: pt.Y})
				pa.Line(pt)
			case PostStep:
				pa.Line(vg.Point{X: pt.X, Y: prev.Y})
				if i != len(ps)-2 {
					pa.Line(pt)
				}
			}
			prev = pt
		}
		pa.Line(vg.Point{X: ps[len(pts.XYs)-1].X, Y: minY})
		pa.Close()
		c.Fill(pa)
	}

	if pts.LineStyle != nil {
		lines := c.ClipLinesXY(ps)
		if len(lines) == 0 {
			return
		}
		c.SetLineStyle(*pts.LineStyle)
		for _, l := range lines {
			if len(l) == 0 {
				continue
			}
			var p vg.Path
			prev := l[0]
			p.Move(prev)
			for _, pt := range l[1:] {
				switch pts.StepStyle {
				case PreStep:
					p.Line(vg.Point{X: prev.X, Y: pt.Y})
				case MidStep:
					p.Line(vg.Point{X: (prev.X + pt.X) / 2, Y: prev.Y})
					p.Line(vg.Point{X: (prev.X + pt.X) / 2, Y: pt.Y})
				case PostStep:
					p.Line(vg.Point{X: pt.X, Y: prev.Y})
				}
				p.Line(pt)
				prev = pt
			}
			c.Stroke(p)
		}
	}
}

// DataRange returns the minimum and maximum
// x and y values, implementing the plot.DataRanger
// interface.
func (pts *Step) DataRange() (xmin, xmax, ymin, ymax float64) {
	if pts.FillColor != nil {
		xmin, xmax, ymin, ymax = XYRange(pts)
		ymin = math.Min(ymin, 0.)
		ymax = math.Max(ymax, 0.)
		return
	}
	return XYRange(pts)
}

// Thumbnail returns the thumbnail for the Step, implementing the plot.Thumbnailer interface.
func (pts *Step) Thumbnail(c *draw.Canvas) {
	if pts.FillColor != nil {
		topY := vg.Length(0.)
		if pts.LineStyle == nil {
			topY = c.Max.Y
		} else {
			topY = (c.Min.Y + c.Max.Y) / 2
		}
		points := []vg.Point{
			{X: c.Min.X, Y: c.Min.Y},
			{X: c.Min.X, Y: topY},
			{X: c.Max.X, Y: topY},
			{X: c.Max.X, Y: c.Min.Y},
		}
		poly := c.ClipPolygonY(points)
		c.FillPolygon(pts.FillColor, poly)
	}

	if pts.LineStyle != nil {
		y := c.Center().Y
		c.StrokeLine2(*pts.LineStyle, c.Min.X, y, c.Max.X, y)
	}
}