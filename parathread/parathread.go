package parathread

import (
	"fmt"
	"github.com/aeud/goutils/firelog"
	"log"
	"regexp"
	"strings"
	"sync"
)

type Node struct {
	Key      string
	KeyDeps  []string
	Deps     []*Node
	Executed *sync.WaitGroup
}

func newNode(k string, ks []string) *Node {
	n := new(Node)
	n.Key = k
	n.KeyDeps = ks
	n.Executed = new(sync.WaitGroup)
	n.Executed.Add(1)
	return n
}

func (n *Node) Run(f func(n *Node)) {
	for _, c := range n.Deps {
		c.Executed.Wait()
	}
	f(n)
	n.Executed.Done()
}

type MapNodes map[string]*Node

func (m MapNodes) String() string {
	s := make([]string, len(m))
	for k, i := range m {
		ds := make([]string, len(i.Deps))
		for i, n := range i.Deps {
			ds[i] = n.Key
		}
		s = append(s, fmt.Sprintf("%v depends on %v nodes (%v)", k, len(i.Deps), strings.Join(ds, ", ")))
	}
	return strings.Join(s, "\n")
}

type Thread struct {
	Closed bool
	Map    MapNodes
	wg     *sync.WaitGroup
	logger *firelog.FirebaseService
}

func NewThread() *Thread {
	t := new(Thread)
	t.Closed = false
	t.Map = make(MapNodes)
	t.wg = new(sync.WaitGroup)
	return t
}

func (t *Thread) AddLogger(endpoint, authToken, ref string) {
	t.logger = firelog.NewFirebaseService(endpoint, authToken, ref)
	t.logger.Deamon()
}

func (t *Thread) Log(k, c string) {
	if t.logger != nil {
		t.logger.Push(firelog.NewFirebaseMessage(k, c))
	}
}

func (t *Thread) String() string {
	layout := `
Length: %v
Closed: %v
Keys: %v
`
	return fmt.Sprintf(layout, len(t.Map), t.Closed, t.Map)
}

func (t *Thread) Add(k string, ks []string) {
	if !t.Closed {
		t.Map[k] = newNode(k, ks)
	} else {
		panic("Closed thread")
	}
}

func (t *Thread) Close() {
	t.Closed = true
}

func (t *Thread) Prepare() {
	t.Close()
	for _, v := range t.Map {
		kd := removeDuplicate(v.KeyDeps)
		v.Deps = make([]*Node, 0)
		for _, key := range kd {
			if t.Map[key] != nil {
				v.Deps = append(v.Deps, t.Map[key])
			} else {
				log.Printf("Ignore dep %v\n", key)
			}
		}
	}
}

func (t *Thread) Run(f func(n *Node)) {
	for _, n := range t.Map {
		t.wg.Add(1)
		t.Log(n.Key, "pending...")
		go func(n *Node, t *Thread) {
			n.Run(f)
			t.Log(n.Key, "done")
			t.wg.Done()
		}(n, t)
	}
	t.wg.Wait()
	if t.logger != nil {
		t.logger.Wait()
	}
}

func (t *Thread) RebuildFromRegexp(pattern string) *Thread {
	if !t.Closed {
		t.Prepare()
	}
	nt := NewThread()
	for k := range t.Map {
		matched, err := regexp.MatchString(pattern, k)
		if err != nil {
			log.Fatal(err)
		}
		if matched {
			nt.RecBuildFromNode(t.Map[k])
		}
	}
	nt.Prepare()
	return nt
}

func (t *Thread) RebuildFromKey(k string) *Thread {
	if !t.Closed {
		t.Prepare()
	}
	nt := NewThread()
	nt.RecBuildFromNode(t.Map[k])
	nt.Prepare()
	return nt
}

func (t *Thread) RecBuildFromNode(n *Node) {
	t.Add(n.Key, n.KeyDeps)
	for _, c := range n.Deps {
		t.RecBuildFromNode(c)
	}
}

func removeDuplicate(a []string) []string {
	mem := make(map[string]int)
	for _, v := range a {
		mem[v]++
	}
	newArray := make([]string, len(mem))
	var i int
	for k := range mem {
		newArray[i] = k
		i++
	}
	return newArray
}
