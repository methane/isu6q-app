package main

import (
	"crypto/sha1"
	"fmt"
	"html"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

type Keyword struct {
	Key    string
	Link   string
	holder string
}

type KeywordArray []Keyword

var (
	mKwControl sync.Mutex
	kwdList    KeywordArray

	mUpdateReplacer sync.Mutex
	repLastUpdated  time.Time

	mKwReplacer                  sync.RWMutex
	kwReplacer1st, kwReplacer2nd *strings.Replacer
)

func updateReplacer() {
	now := time.Now()
	mUpdateReplacer.Lock()
	defer mUpdateReplacer.Unlock()

	if repLastUpdated.After(now) {
		return
	}
	repLastUpdated = time.Now()

	reps1 := make([]string, 0, len(kwdList)*2)
	reps2 := make([]string, 0, len(kwdList)*2)

	mKwControl.Lock()
	kws := kwdList[:]
	mKwControl.Unlock()
	sort.Sort(kws)

	for i, k := range kwdList {
		if k.holder == "" {
			k.holder = fmt.Sprintf("isuda_%x", sha1.Sum([]byte(k.Key)))
			kwdList[i].holder = k.holder
		}

		reps1 = append(reps1, k.Key)
		reps1 = append(reps1, k.holder)

		reps2 = append(reps2, k.holder)
		reps2 = append(reps2, k.Link)
	}
	r1 := strings.NewReplacer(reps1...)
	r2 := strings.NewReplacer(reps2...)
	mKwReplacer.Lock()
	kwReplacer1st = r1
	kwReplacer2nd = r2
	mKwReplacer.Unlock()
}

func AddKeyword(key, link string) {
	k := Keyword{Key: key, Link: link}

	mKwControl.Lock()
	kwdList = append(kwdList, k)
	mKwControl.Unlock()

	updateReplacer()
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
	mKwControl.Unlock()

	updateReplacer()
}

func InitKeyword(kws KeywordArray) {
	mKwControl.Lock()
	kwdList = kws
	mKwControl.Unlock()
	updateReplacer()
}

func ReplaceKeyword(c string) string {
	mKwReplacer.RLock()
	r1 := kwReplacer1st
	r2 := kwReplacer2nd
	mKwReplacer.RUnlock()
	x := r1.Replace(c)
	x = html.EscapeString(x)
	return r2.Replace(x)
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
