// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plotter

import (
	"image/color"
	"math"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// StepType specifies a form of a connection of two consecutive points.
type StepType int

const (
	// StepTypePre means that two consecutive points are connected by two lines in the following order: vertical, horizontal.
	StepTypePre StepType = iota

	// StepTypeMid means that two consecutive points are connected by three lines in the following order: horizontal, vertical, horizontal.
	// Vertical line is placed in the middle of the interval.
	StepTypeMid

	// StepTypePost means that two consecutive points are connected by two lines in the following order: horizontal, vertical.
	StepTypePost
)

// Step implements the Plotter interface, drawing a stepped line.
type Step struct {
	// XYs is a copy of the points for this line.
	XYs

	// StepStyle is the type of step line
	StepType StepType

	// LineStyle is the style of the line connecting the points.
	LineStyle *draw.LineStyle

	// ShadeColor is the color of the shaded area.
	ShadeColor color.Color
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

	if pts.ShadeColor != nil && len(ps) > 0 {
		c.SetColor(pts.ShadeColor)
		minY := trY(plt.Y.Min)
		var pa vg.Path
		pa.Move(vg.Point{X: ps[0].X, Y: minY})
		prev := ps[0]
		if pts.StepType != StepTypePre {
			pa.Line(prev)
		}
		for i, pt := range ps[1:] {
			switch pts.StepType {
			case StepTypePre:
				pa.Line(vg.Point{X: prev.X, Y: pt.Y})
				pa.Line(pt)
			case StepTypeMid:
				pa.Line(vg.Point{X: (prev.X + pt.X) / 2, Y: prev.Y})
				pa.Line(vg.Point{X: (prev.X + pt.X) / 2, Y: pt.Y})
				pa.Line(pt)
			case StepTypePost:
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
				switch pts.StepType {
				case StepTypePre:
					p.Line(vg.Point{X: prev.X, Y: pt.Y})
				case StepTypeMid:
					p.Line(vg.Point{X: (prev.X + pt.X) / 2, Y: prev.Y})
					p.Line(vg.Point{X: (prev.X + pt.X) / 2, Y: pt.Y})
				case StepTypePost:
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
	if pts.ShadeColor != nil {
		xmin, xmax, ymin, ymax = XYRange(pts)
		ymin = math.Min(ymin, 0.)
		ymax = math.Max(ymax, 0.)
		return
	}
	return XYRange(pts)
}

// Thumbnail the thumbnail for the Step,
// implementing the plot.Thumbnailer interface.
func (pts *Step) Thumbnail(c *draw.Canvas) {
	if pts.ShadeColor != nil {
		points := []vg.Point{
			{X: c.Min.X, Y: c.Min.Y},
			{X: c.Min.X, Y: c.Max.Y},
			{X: c.Max.X, Y: c.Max.Y},
			{X: c.Max.X, Y: c.Min.Y},
		}
		poly := c.ClipPolygonY(points)
		c.FillPolygon(pts.ShadeColor, poly)
	} else if pts.LineStyle != nil {
		y := c.Center().Y
		c.StrokeLine2(*pts.LineStyle, c.Min.X, y, c.Max.X, y)
	}
}

// NewStepPoints returns both a Step and a
// Points for the given point data.
func NewStepPoints(xys XYer) (*Step, *Scatter, error) {
	s, err := NewScatter(xys)
	if err != nil {
		return nil, nil, err
	}
	l := &Step{
		XYs:       s.XYs,
		LineStyle: &DefaultLineStyle,
	}
	return l, s, nil
}
