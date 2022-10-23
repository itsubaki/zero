package neu

import (
	"fmt"

	"github.com/itsubaki/neu/layer"
	"github.com/itsubaki/neu/math/matrix"
	"github.com/itsubaki/neu/math/numerical"
	"github.com/itsubaki/neu/optimizer"
	"github.com/itsubaki/neu/trainer"
	"github.com/itsubaki/neu/weight"
)

var (
	_ Layer         = (*layer.Add)(nil)
	_ Layer         = (*layer.Affine)(nil)
	_ Layer         = (*layer.Dot)(nil)
	_ Layer         = (*layer.Dropout)(nil)
	_ Layer         = (*layer.Mul)(nil)
	_ Layer         = (*layer.ReLU)(nil)
	_ Layer         = (*layer.Sigmoid)(nil)
	_ Layer         = (*layer.SoftmaxWithLoss)(nil)
	_ Optimizer     = (*optimizer.AdaGrad)(nil)
	_ Optimizer     = (*optimizer.Momentum)(nil)
	_ Optimizer     = (*optimizer.SGD)(nil)
	_ WeightInit    = weight.Std(0.01)
	_ WeightInit    = weight.He
	_ WeightInit    = weight.Xavier
	_ WeightInit    = weight.Glorot
	_ trainer.Model = (*Neu)(nil)
)

type Layer interface {
	Forward(x, y matrix.Matrix) matrix.Matrix
	Backward(dout matrix.Matrix) (matrix.Matrix, matrix.Matrix)
}

type Optimizer interface {
	Update(params, grads map[string]matrix.Matrix) map[string]matrix.Matrix
}

type WeightInit func(prevNodeNum int) float64

type Config struct {
	InputSize         int
	HiddenSize        []int
	OutputSize        int
	WeightDecayLambda float64
	WeightInit        WeightInit
	Optimizer         Optimizer
}

type Neu struct {
	size              []int
	params            map[string]matrix.Matrix
	layer             []Layer
	last              Layer
	weightDecayLambda float64
	optimizer         Optimizer
}

func New(c *Config) *Neu {
	// size
	size := append([]int{c.InputSize}, c.HiddenSize...)
	size = append(size, c.OutputSize)

	// params
	params := make(map[string]matrix.Matrix)
	for i := 0; i < len(size)-1; i++ {
		params[fmt.Sprintf("W%v", i+1)] = matrix.Randn(size[i], size[i+1])
		params[fmt.Sprintf("B%v", i+1)] = matrix.Zero(1, size[i+1])
	}

	// weight init
	for i := 0; i < len(size)-1; i++ {
		W := fmt.Sprintf("W%v", i+1)
		params[W] = matrix.Func(params[W], func(v float64) float64 {
			return c.WeightInit(size[i]) * v
		})
	}

	// new
	return &Neu{
		size:              size,
		params:            params,
		layer:             make([]Layer, 0),
		last:              &layer.SoftmaxWithLoss{},
		weightDecayLambda: c.WeightDecayLambda,
		optimizer:         c.Optimizer,
	}
}

func (n *Neu) Predict(x matrix.Matrix) matrix.Matrix {
	n.layer = make([]Layer, 0) // init
	for i := 0; i < len(n.size)-1; i++ {
		n.layer = append(n.layer, &layer.Affine{
			W: n.params[fmt.Sprintf("W%v", i+1)],
			B: n.params[fmt.Sprintf("B%v", i+1)],
		})
		n.layer = append(n.layer, &layer.ReLU{})
	}
	n.layer = n.layer[:len(n.layer)-1] // remove last ReLU

	for _, l := range n.layer {
		x = l.Forward(x, nil)
	}

	return x
}

func (n *Neu) Loss(x, t matrix.Matrix) matrix.Matrix {
	y := n.Predict(x)
	loss := n.last.Forward(y, t)

	// decay
	var decay float64
	for i := 0; i < len(n.size)-1; i++ {
		sumw2 := n.params[fmt.Sprintf("W%v", i+1)].Func(func(v float64) float64 {
			return v * v
		}).Sum() // sum(W**2)
		decay = decay + 0.5*n.weightDecayLambda*sumw2 // 1/2 * lambda * sum(W**2)
	}

	return loss.Func(func(v float64) float64 { return v + decay })
}

func (n *Neu) Gradient(x, t matrix.Matrix) map[string]matrix.Matrix {
	// forward
	n.Loss(x, t)

	// backward
	dout, _ := n.last.Backward(matrix.New([]float64{1}))
	for i := len(n.layer) - 1; i > -1; i-- {
		dout, _ = n.layer[i].Backward(dout)
	}

	// gradient
	grads := make(map[string]matrix.Matrix)
	var i int
	for j := 0; j < len(n.layer); j++ {
		if _, ok := n.layer[j].(*layer.Affine); !ok {
			continue
		}

		grads[fmt.Sprintf("W%v", i+1)] = n.layer[j].(*layer.Affine).DW
		grads[fmt.Sprintf("B%v", i+1)] = n.layer[j].(*layer.Affine).DB
		i++
	}

	// decay
	for i := 0; i < len(n.size)-1; i++ {
		W := fmt.Sprintf("W%v", i+1)
		grads[W] = grads[W].FuncWith(n.params[W], func(a, b float64) float64 {
			return a + n.weightDecayLambda*b // grads[W] + lambda * params[W]
		})
	}

	return grads
}

func (n *Neu) NumericalGradient(x, t matrix.Matrix) map[string]matrix.Matrix {
	lossW := func(w ...float64) float64 {
		return n.Loss(x, t)[0][0]
	}

	grad := func(f func(x ...float64) float64, x matrix.Matrix) matrix.Matrix {
		out := make(matrix.Matrix, 0)
		for _, r := range x {
			out = append(out, numerical.Gradient(f, r))
		}

		return out
	}

	// gradient
	grads := make(map[string]matrix.Matrix)
	for i := 0; i < len(n.size)-1; i++ {
		W, B := fmt.Sprintf("W%v", i+1), fmt.Sprintf("B%v", i+1)
		grads[W] = grad(lossW, n.params[W])
		grads[B] = grad(lossW, n.params[B])
	}

	// decay
	for i := 0; i < len(n.size)-1; i++ {
		W := fmt.Sprintf("W%v", i+1)
		grads[W] = grads[W].FuncWith(n.params[W], func(a, b float64) float64 {
			return a + n.weightDecayLambda*b // grads[W] + lambda * params[W]
		})
	}

	return grads
}

func (n *Neu) Optimize(grads map[string]matrix.Matrix) {
	n.params = n.optimizer.Update(n.params, grads)
}
