package plot

import (
	"fmt"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func XRange(begin, end, delta float64) []float64 {
	out := []float64{begin}
	for {
		begin = begin + delta
		if begin > end {
			break
		}

		out = append(out, begin)
	}

	return out
}

func Save(xrange, yrange []float64, filename string) error {
	xys := make(plotter.XYs, 0)
	for i := range xrange {
		xys = append(xys, plotter.XY{
			X: xrange[i],
			Y: yrange[i],
		})
	}

	line, err := plotter.NewLine(xys)
	if err != nil {
		return fmt.Errorf("plotter newline: %v", err)
	}

	p := plot.New()
	p.Add(line)

	if err := p.Save(4*vg.Inch, 4*vg.Inch, filename); err != nil {
		return fmt.Errorf("save: %v", err)
	}

	return nil
}
