// Copyright ©2012 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vgimg_test

import (
	"bytes"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/igrmk/plot"
	"github.com/igrmk/plot/plotter"
	"github.com/igrmk/plot/vg"
	"github.com/igrmk/plot/vg/draw"
	"github.com/igrmk/plot/vg/vgimg"
)

func TestIssue179(t *testing.T) {
	scatter, err := plotter.NewScatter(plotter.XYs{{1, 1}, {0, 1}, {0, 0}})
	if err != nil {
		log.Fatal(err)
	}
	p, err := plot.New()
	if err != nil {
		log.Fatal(err)
	}
	p.Add(scatter)
	p.HideAxes()

	c := vgimg.JpegCanvas{Canvas: vgimg.New(5.08*vg.Centimeter, 5.08*vg.Centimeter)}
	p.Draw(draw.New(c))
	b := bytes.NewBuffer([]byte{})
	if _, err = c.WriteTo(b); err != nil {
		t.Error(err)
	}

	f, err := os.Open(filepath.Join("testdata", "issue179.jpg"))
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	want, err := ioutil.ReadAll(f)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(b.Bytes(), want) {
		t.Error("Image mismatch")
	}
}

func TestConcurrentInit(t *testing.T) {
	vg.MakeFont("Helvetica", 10)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		c := vgimg.New(215, 215)
		c.FillString(vg.Font{Size: 10}, vg.Point{}, "hi")
		wg.Done()
	}()
	go func() {
		c := vgimg.New(215, 215)
		c.FillString(vg.Font{Size: 10}, vg.Point{}, "hi")
		wg.Done()
	}()
	wg.Wait()
}

func TestUseBackgroundColor(t *testing.T) {
	colors := []color.Color{color.Transparent, color.NRGBA{R: 255, A: 255}}
	for i, col := range colors {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			c := vgimg.NewWith(vgimg.UseWH(1, 1), vgimg.UseBackgroundColor(col))
			img := c.Image()
			wantCol := color.RGBAModel.Convert(col)
			haveCol := img.At(0, 0)
			if !reflect.DeepEqual(haveCol, wantCol) {
				t.Fatalf("color should be %#v but is %#v", wantCol, haveCol)
			}
		})
	}
}
