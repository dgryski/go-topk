package topk

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"log"
	"os"
	"reflect"
	"sort"
	"testing"
)

type freqs struct {
	keys   []string
	counts map[string]int
}

func (f freqs) Len() int { return len(f.keys) }

// Actually 'Greater', since we want decreasing
func (f *freqs) Less(i, j int) bool {
	return f.counts[f.keys[i]] > f.counts[f.keys[j]] || f.counts[f.keys[i]] == f.counts[f.keys[j]] && f.keys[i] < f.keys[j]
}

func (f *freqs) Swap(i, j int) { f.keys[i], f.keys[j] = f.keys[j], f.keys[i] }

func TestTopK(t *testing.T) {

	f, err := os.Open("testdata/domains.txt")

	if err != nil {
		t.Fatal(err)
	}

	scanner := bufio.NewScanner(f)

	tk := New(100)
	exact := make(map[string]int)

	for scanner.Scan() {

		item := scanner.Text()

		exact[item]++
		e := tk.Insert(item, 1)
		if e.Count < exact[item] {
			t.Errorf("estimate lower than exact: key=%v, exact=%v, estimate=%v", e.Key, exact[item], e.Count)
		}
		if e.Count-e.Error > exact[item] {
			t.Errorf("error bounds too large: key=%v, count=%v, error=%v, exact=%v", e.Key, e.Count, e.Error, exact[item])
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("error during scan: ", err)
	}

	var keys []string

	for k, _ := range exact {
		keys = append(keys, k)
	}

	freq := &freqs{keys: keys, counts: exact}

	sort.Sort(freq)

	top := tk.Keys()

	// at least the top 25 must be in order
	for i := 0; i < 25; i++ {
		if top[i].Key != freq.keys[i] {
			t.Errorf("key mismatch: idx=%d top=%s (%d) exact=%s (%d)", i, top[i].Key, top[i].Count, freq.keys[i], freq.counts[freq.keys[i]])
		}
	}
	for k, v := range exact {
		e := tk.Estimate(k)
		if e.Count < v {
			t.Errorf("estimate lower than exact: key=%v, exact=%v, estimate=%v", e.Key, v, e.Count)
		}
		if e.Count-e.Error > v {
			t.Errorf("error bounds too large: key=%v, count=%v, error=%v, exact=%v", e.Key, e.Count, e.Error, v)
		}
	}
	for _, k := range top {
		e := tk.Estimate(k.Key)
		if e != k {
			t.Errorf("estimate differs from top keys: key=%v, estimate=%v(-%v) top=%v(-v)", e.Key, e.Count, e.Error, k.Count, k.Error)
		}
	}

	// gob
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(tk); err != nil {
		t.Error(err)
	}

	decoded := New(100)
	dec := gob.NewDecoder(&buf)
	if err := dec.Decode(decoded); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(tk, decoded) {
		t.Error("they are not equal.")
	}
}
