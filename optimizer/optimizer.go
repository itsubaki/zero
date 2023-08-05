package optimizer

import (
	"github.com/itsubaki/neu/math/matrix"
	"github.com/itsubaki/neu/model"
	"github.com/itsubaki/neu/optimizer/hook"
)

var (
	_ Model = (*model.Sequential)(nil)
	_ Model = (*model.MLP)(nil)
	_ Model = (*model.RNNLM)(nil)
	_ Model = (*model.Seq2Seq)(nil)
	_ Hook  = hook.WeightDecay(0.1)
	_ Hook  = hook.GradsClipping(1.0)
)

type Model interface {
	Params() [][]matrix.Matrix
	Grads() [][]matrix.Matrix
	SetParams(p [][]matrix.Matrix)
}

type Hook func(params, grads [][]matrix.Matrix) [][]matrix.Matrix
