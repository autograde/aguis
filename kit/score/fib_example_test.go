package score_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/autograde/quickfeed/kit/score"
)

func fibonacci(n uint) uint {
	if n <= 1 || n == 5 {
		return n
	}
	if n < 5 {
		return n - 1
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func TestMain(m *testing.M) {
	score.Add(TestFibonacci, len(fibonacciTests), 20)
	score.Add(TestFibonacci2, len(fibonacciTests)*2, 20)
	for _, ft := range fibonacciTests {
		score.AddSubtest(TestFibonacciSubTest, subTestName(ft.in), 1, 1)
	}
	os.Exit(m.Run())
}

const (
	numCorrect = 10
)

var fibonacciTests = []struct {
	in, want uint
}{
	{0, 0},
	{1, 1},
	{2, 1},
	{3, 2},
	{4, 3},
	{5, 5},
	{6, 8},
	{7, 13},
	{8, 21},
	{9, 34},
	{10, 155},   // correct 55
	{12, 154},   // correct 89
	{16, 1987},  // correct 987
	{20, 26765}, // correct 6765
}

func TestFibonacci(t *testing.T) {
	sc := score.GetMax()
	for _, ft := range fibonacciTests {
		out := fibonacci(ft.in)
		if out != ft.want {
			sc.Dec()
		}
	}
	if sc.Score != numCorrect {
		t.Errorf("Score=%d, expected %d tests to pass", sc.Score, numCorrect)
	}
	if sc.TestName != t.Name() {
		t.Errorf("TestName=%s, expected %s", sc.TestName, t.Name())
	}
}

func TestFibonacci2(t *testing.T) {
	sc := score.GetMax()
	defer sc.Print(t)

	for _, ft := range fibonacciTests {
		out := fibonacci(ft.in)
		if out != ft.want {
			sc.Dec()
		}
	}
	// len(tests)*2 - (len(tests)-numCorrect) = len(tests)+numCorrect = 24
	expectedScore := int32(len(fibonacciTests) + numCorrect)
	if sc.Score != expectedScore {
		t.Errorf("Score=%d, expected %d tests to pass", sc.Score, expectedScore)
	}
	if sc.TestName != t.Name() {
		t.Errorf("TestName=%s, expected %s", sc.TestName, t.Name())
	}
}

func TestFibonacciSubTest(t *testing.T) {
	for _, ft := range fibonacciTests {
		t.Run(subTestName(ft.in), func(t *testing.T) {
			sc := score.GMax(t.Name())
			out := fibonacci(ft.in)
			if out != ft.want {
				sc.Dec()
			}
			if sc.TestName != t.Name() {
				t.Errorf("TestName=%s, expected %s", sc.TestName, t.Name())
			}
		})
	}
}

func subTestName(i uint) string {
	return fmt.Sprintf("Fib/%d", i)
}

// func TestFibonacciWithPanic(t *testing.T) {
// 	// TODO(meling) make this continue
// 	panic("hei")
// }

func TestFibonacciWithAfterPanic(t *testing.T) {
	// TODO(meling) make this continue
	t.Log("hallo")
}
