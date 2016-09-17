package main

import (
	"log"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"
)

type Keyword struct {
	Key  string
	Link string
}

type KeywordArray []Keyword

var (
	mKwControl sync.Mutex
	kwdList    KeywordArray

	mKwReplacer sync.RWMutex
	kwReplacer  *strings.Replacer
)

func updateReplacer() {
	reps := make([]string, 0, len(kwdList)*2)
	for _, k := range kwdList {
		reps = append(reps, k.Key)
		reps = append(reps, k.Link)
	}
	rl := strings.NewReplacer(reps...)
	mKwReplacer.Lock()
	kwReplacer = rl
	mKwReplacer.Unlock()
}

func AddKeyword(key, link string) {
	k := Keyword{key, link}

	mKwControl.Lock()
	kwdList = append(kwdList, k)
	sort.Sort(kwdList)

	updateReplacer()
	mKwControl.Unlock()
}

func RemoveKeyword(key string) {
	mKwControl.Lock()
	var pos int = -1
	for i, k := range kwdList {
		if k.Key == key {
			pos = i
			break
		}
	}
	if pos < 0 {
		log.Printf("keyword not found: %q", key)
		return
	}
	kwdList[pos] = kwdList[len(kwdList)-1]
	kwdList = kwdList[:len(kwdList)-1]
	sort.Sort(kwdList)

	updateReplacer()
	mKwControl.Unlock()
}

func InitKeyword(kws KeywordArray) {
	mKwControl.Lock()
	kwdList = kws
	sort.Sort(kwdList)

	updateReplacer()
	mKwControl.Unlock()
}

func ReplaceKeyword(c string) string {
	mKwReplacer.RLock()
	rp := kwReplacer
	mKwReplacer.RUnlock()
	return rp.Replace(c)
}

var _ sort.Interface = KeywordArray{}

func (ks KeywordArray) Len() int {
	return len(ks)
}

func (ks KeywordArray) Less(i, j int) bool {
	return utf8.RuneCountInString(ks[i].Key) > utf8.RuneCountInString(ks[j].Key)
}

func (ks KeywordArray) Swap(i, j int) {
	ks[i], ks[j] = ks[j], ks[i]
}
