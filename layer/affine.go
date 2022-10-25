package layer

import "github.com/itsubaki/neu/math/matrix"

type Affine struct {
	W  matrix.Matrix
	B  matrix.Matrix
	x  matrix.Matrix
	DW matrix.Matrix
	DB matrix.Matrix
}

func (l *Affine) Forward(x, _ matrix.Matrix, opts ...Opts) matrix.Matrix {
	l.x = x
	return matrix.Dot(l.x, l.W).Add(l.B) // x.W + b
}

func (l *Affine) Backward(dout matrix.Matrix) (matrix.Matrix, matrix.Matrix) {
	dx := matrix.Dot(dout, l.W.T())
	l.DW = matrix.Dot(l.x.T(), dout)
	l.DB = SumAxis0(dout)
	return dx, matrix.New()
}

func SumAxis0(m matrix.Matrix) matrix.Matrix {
	p, q := m.Dimension()

	v := make([]float64, q)
	for i := 0; i < q; i++ {
		for j := 0; j < p; j++ {
			v[i] = v[i] + m[j][i]
		}
	}

	return matrix.New(v)
}
