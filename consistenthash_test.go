package consistenthash

import (
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"math/rand"
	"sort"
	"strconv"
	"testing"
)

func TestBuildCH(t *testing.T) {
	fmt.Printf("\n---\nTEST BUILD, INSERT AND FIND\n")
	mult := 100
	ch, _ := NewConsistentHash(mult)
	items := []string{"127.0.0.1", "17.0.1.1", "1.1.0.1", "27.99.0.111", "64.0.8.8", "8.8.8.8", "10.100.0.100",
		"128.4.4.4", "28.28.1.1", "28.10.0.10", "12.9.0.10", "11.11.8.1", "13.10.0.19", "128.19.19.19"}
	for _, item := range items {
		insertErr := ch.Insert(item)
		if insertErr != nil {
			t.Errorf(insertErr.Error())
		}
	}
	// make sure the len is correct
	if len(ch.SumList) != (mult * len(items)) {
		e := fmt.Sprintf("SumList len should be %d, but is %d", (mult * len(items)), len(ch.SumList))
		t.Errorf(e)
	}
	// make sure the list is indeed sorted
	sl := make([]int, len(ch.SumList))
	for k, v := range ch.SumList {
		sl[k] = int(v)
	}
	sort.Ints(sl)
	for k, v := range sl {
		if v != int(ch.SumList[k]) {
			e := fmt.Sprintf("sorted list differs at position %d", k)
			t.Errorf(e)
		}
	}
	count := make(map[string]int)
	for _, sum := range ch.SumList {
		count[ch.Source[sum]]++
	}
	for k, c := range count {
		if c != mult {
			e := fmt.Sprintf("%s appears %d times, not %d times", k, c, mult)
			t.Errorf(e)
		}
	}
	nearestHash, nhErr := ch.Find("hello")
	if nhErr != nil || nearestHash == "" {
		t.Errorf("err returned on finding nearest hash for hello")
	}

	// test distribution
	dist := make(map[string]int)
	total := 10000
	for j := 0; j < total; j++ {
		b64 := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(int(rand.Int31()))))
		nearestHash, nhErr := ch.Find(b64)
		if nhErr != nil || nearestHash == "" {
			e := fmt.Sprintf("err returned on finding nearest hash for %s", b64)
			t.Errorf(e)
		}
		dist[nearestHash]++
	}
	for k, v := range dist {
		fmt.Printf("%s %d (%f pct) \n", k, v, (float64(v)/float64(total))*100)
	}
}

func TestEmptyCH(t *testing.T) {
	fmt.Printf("\n---\nTEST EMPTY\n")
	ch, chErr := NewConsistentHash(0)
	if chErr == nil {
		t.Errorf("should return nil when mult factor 0")
	}
	ch, chErr = NewConsistentHash(1)
	if chErr != nil {
		t.Errorf("should return non-nil when mult factor nonzero")
	}
	_, err := ch.Find("hello")
	if err == nil {
		t.Errorf("should not find anything on empty hash")
	}
}

func TestCollision(t *testing.T) {
	fmt.Printf("\n---\nTEST COLLISION\n")
	ch, _ := NewConsistentHash(1)
	insertErr := ch.Insert("hello")
	if insertErr != nil {
		t.Errorf(insertErr.Error())
	}
	insertErr = ch.Insert("hello")
	if insertErr == nil {
		t.Errorf("should have caused collision")
	}
}

func TestRemove(t *testing.T) {
	fmt.Printf("\n---\nTEST REMOVE\n")
	mult := 2
	ch, _ := NewConsistentHash(mult)
	items := []string{"127.0.0.1", "17.0.1.1", "1.1.0.1", "27.99.0.111", "64.0.8.8", "8.8.8.8", "10.100.0.100",
		"128.4.4.4", "28.28.1.1", "28.10.0.10", "12.9.0.10", "11.11.8.1", "13.10.0.19", "128.19.19.19"}
	for _, item := range items {
		insertErr := ch.Insert(item)
		if insertErr != nil {
			t.Errorf(insertErr.Error())
		}
	}

	// remove something that isn't there
	_ = ch.Remove("hello")

	// make sure the len is correct
	if len(ch.SumList) != (mult * len(items)) {
		e := fmt.Sprintf("SumList len should be %d, but is %d", (mult * len(items)), len(ch.SumList))
		t.Errorf(e)
	}

	// remove something that is there
	_ = ch.Remove("28.28.1.1")

	// make sure the len is correct
	if len(ch.SumList) != (mult * (len(items) - 1)) {
		e := fmt.Sprintf("SumList len should be %d, but is %d", (mult * len(items)), len(ch.SumList))
		t.Errorf(e)
	}

	sum1 := crc32.ChecksumIEEE([]byte(multElt("28.28.1.1", 1)))
	sum2 := crc32.ChecksumIEEE([]byte(multElt("28.28.1.1", 2)))
	for _, v := range ch.SumList {
		if v == sum1 || v == sum2 {
			t.Errorf("found element in SumList that should have been deleted")
		}
	}

	mult = 1
	ch, _ = NewConsistentHash(mult)
	items = []string{"127.0.0.1", "17.0.1.1", "1.1.0.1", "27.99.0.111", "64.0.8.8", "8.8.8.8", "10.100.0.100",
		"128.4.4.4", "28.28.1.1", "28.10.0.10", "12.9.0.10", "11.11.8.1", "13.10.0.19", "128.19.19.19"}
	for _, item := range items {
		insertErr := ch.Insert(item)
		if insertErr != nil {
			t.Errorf(insertErr.Error())
		}
	}

	// find the target for a source. then remove that target. then reinsert it and make sure
	// source maps once to it once again
	nearest, err := ch.Find("helloooo")
	if err != nil {
		t.Errorf(err.Error())
	}
	fmt.Printf("initially source maps to target %s\n", nearest)
	_ = ch.Remove(nearest)
	nearest2, err2 := ch.Find("helloooo")
	if err2 != nil {
		t.Errorf(err2.Error())
	}
	fmt.Printf("after removal of %s, source now maps to %s\n", nearest, nearest2)
	insertErr := ch.Insert(nearest)
	if insertErr != nil {
		t.Errorf(insertErr.Error())
	}
	nearest3, err3 := ch.Find("helloooo")
	if err3 != nil {
		t.Errorf(err3.Error())
	}
	fmt.Printf("after reinsertion of %s, source now maps to %s\n", nearest, nearest3)
	if nearest != nearest3 {
		t.Errorf("was not able to round trip removal and addition of target")
	}
}
