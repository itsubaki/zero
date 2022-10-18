package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/itsubaki/neu"
	"github.com/itsubaki/neu/math/matrix"
	"github.com/itsubaki/neu/mnist"
	"github.com/itsubaki/neu/optimizer"
)

func main() {
	var dir string
	flag.StringVar(&dir, "dir", "./testdata", "")
	flag.Parse()

	// data
	train, test := mnist.Must(mnist.Load(dir))

	x := matrix.New(mnist.Normalize(train.Image)...)
	t := matrix.New(mnist.OneHot(train.Label)...)

	xt := matrix.New(mnist.Normalize(test.Image)...)
	tt := matrix.New(mnist.OneHot(test.Label)...)

	// hyper-parameter
	batchSize := 100
	iter := 10000

	// init
	rand.Seed(time.Now().Unix())
	n := neu.New(&neu.Config{
		InputSize:         784, // 24 * 24
		HiddenSize:        []int{50, 50, 50},
		OutputSize:        10, // 0 ~ 9
		WeightDecayLambda: 1e-6,
		WeightInit:        neu.He,
		Optimizer:         &optimizer.AdaGrad{LearningRate: 0.01},
	})

	// learning
	for i := 0; i < iter+1; i++ {
		mask := neu.Random(train.N, batchSize)
		xbatch := neu.Batch(x, mask)
		tbatch := neu.Batch(t, mask)

		grads := n.Gradient(xbatch, tbatch)
		n.Optimize(grads)

		if i%(iter/batchSize) == 0 {
			// train data
			loss := n.Loss(xbatch, tbatch)
			acc := neu.Accuracy(n.Predict(xbatch), tbatch)

			// test data
			mask := neu.Random(test.N, batchSize)
			xtbatch := neu.Batch(xt, mask)
			ttbatch := neu.Batch(tt, mask)
			yt := n.Predict(xtbatch)
			tacc := neu.Accuracy(yt, ttbatch)

			// print
			fmt.Printf("%4d: loss=%.04f, train_acc=%.04f, test_acc=%.04f\n", i, loss, acc, tacc)
			fmt.Printf("predict: %v\n", yt.Argmax()[:20])
			fmt.Printf("label  : %v\n", ttbatch.Argmax()[:20])
			fmt.Println()
		}
	}
}
