package gencache

import (
	"github.com/distrue/gencache/src/server/util"
)

const TOP_N = 10
const MAX_COUNT_ENTRY = 99999

var Counter = make(map[string]int, 999999)

func CountItem(key string) {
	Counter[key] += 1
}

func TopN() []util.Node {
	maxHeap := util.NewMaxHeap(99999)
	for key, item := range Counter {
		put := util.Node{
			Val: item,
			Id:  key,
		}
		maxHeap.Insert(put)
		Counter[key] = 0
	}
	var ans []util.Node
	for i := 0; i < TOP_N; i++ {
		put := maxHeap.Remove()
		ans = append(ans, put)
	}
	return ans
}
