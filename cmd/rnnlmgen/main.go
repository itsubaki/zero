package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/itsubaki/neu/activation"
	"github.com/itsubaki/neu/dataset/ptb"
	"github.com/itsubaki/neu/math/matrix"
	"github.com/itsubaki/neu/math/vector"
	"github.com/itsubaki/neu/model"
	"github.com/itsubaki/neu/weight"
)

func main() {
	var dir, start string
	var length int
	flag.StringVar(&dir, "dir", "./testdata", "")
	flag.StringVar(&start, "start", "you", "")
	flag.IntVar(&length, "length", 100, "")
	flag.Parse()

	train := ptb.Must(ptb.Load(dir, ptb.TrainTxt))

	m := &RNNMLGen{
		RNNLM: *model.NewRNNLM(&model.RNNLMConfig{
			VocabSize:   vector.Max(train.Corpus) + 1,
			WordVecSize: 100,
			HiddenSize:  100,
			WeightInit:  weight.Xavier,
		}),
	}

	startID := train.WordToID[start]
	skipIDs := []int{train.WordToID["N"], train.WordToID["<unk>"], train.WordToID["$"]}

	wordIDs := m.Generate(startID, skipIDs, length)
	words := make([]string, length)
	for i, id := range wordIDs {
		words[i] = train.IDToWord[id]
	}

	txt := strings.ReplaceAll(strings.Join(words, " "), "<eos>", "\n")
	fmt.Println(txt)
}

type RNNMLGen struct {
	model.RNNLM
}

func (g *RNNMLGen) Generate(startID int, skipIDs []int, length int) []int {
	wordIDs := []int{startID}

	for {
		if len(wordIDs) >= length {
			break
		}

		// predict
		xs := []matrix.Matrix{matrix.New([]float64{float64(startID)})}
		score := g.Predict(xs)
		p := activation.Softmax(Flatten(score))

		// sample
		sampled := Choice(p)
		if Contains(sampled, skipIDs) {
			continue
		}
		wordIDs = append(wordIDs, sampled)

		// next
		startID = sampled
	}

	return wordIDs
}

func Flatten(m []matrix.Matrix) []float64 {
	flatten := make([]float64, 0)
	for _, s := range m {
		flatten = append(flatten, matrix.Flatten(s)...)
	}

	return flatten
}

func Contains(v int, s []int) bool {
	for _, ss := range s {
		if v == ss {
			return true
		}
	}

	return false
}

func Choice(p []float64) int {
	cumsum := make([]float64, len(p))
	var sum float64
	for i, prob := range p {
		sum += prob
		cumsum[i] = sum
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano())).Float64()
	for i, prop := range cumsum {
		if r <= prop {
			return i
		}
	}

	return -1
}
