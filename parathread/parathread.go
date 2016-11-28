package parathread

import (
	"fmt"
	"github.com/aeud/goutils/firelog"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	dryTimeout time.Duration = 2
)

type Node struct {
	Key      string
	KeyDeps  []string
	Deps     []*Node
	Executed *sync.WaitGroup
	Function func() error
}

func newNode(k string, ks []string, f func() error) *Node {
	n := new(Node)
	n.Key = k
	n.KeyDeps = ks
	n.Executed = new(sync.WaitGroup)
	n.Executed.Add(1)
	n.Function = f
	return n
}

func (n *Node) Run() (chan bool, chan error) {
	armed := make(chan bool)
	res := make(chan error)
	go func() {
		for _, c := range n.Deps {
			c.Executed.Wait()
		}
		armed <- true
		err := n.Function()
		n.Executed.Done()
		res <- err
	}()
	return armed, res
}

func (n *Node) DryRun() (chan bool, chan error) {
	armed := make(chan bool)
	res := make(chan error)
	go func() {
		for _, c := range n.Deps {
			c.Executed.Wait()
		}
		armed <- true
		time.Sleep(dryTimeout * time.Second)
		n.Executed.Done()
		res <- nil
	}()
	return armed, res
}

func (n *Node) HasDep(m *Node) bool {
	for _, d := range n.Deps {
		if d.Key == m.Key {
			return true
		}
	}
	return false
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
		t.logger.AsyncPush(firelog.NewFirebaseMessage(k, c))
	}
}

func (t *Thread) DryLog(k, c string) {
	log.Printf("Dry log - %v: %v\n", k, c)
}

func (t *Thread) String() string {
	layout := `
Length: %v
Closed: %v
Keys: %v
`
	return fmt.Sprintf(layout, len(t.Map), t.Closed, t.Map)
}

func (t *Thread) Add(k string, ks []string, f func() error) {
	if !t.Closed {
		t.Map[k] = newNode(k, ks, f)
	} else {
		panic("Closed thread")
	}
}

func (t *Thread) AddNode(n *Node) {
	if !t.Closed {
		t.Map[n.Key] = n
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

func (t *Thread) Run() {
	t.Prepare()
	for _, n := range t.Map {
		t.wg.Add(1)
		t.Log(n.Key, "pending...")
		go func(n *Node, t *Thread) {
			armed, res := n.Run()
			<-armed
			start := time.Now()
			t.Log(n.Key, fmt.Sprintf("executing... (start: %v)", start))
			if err := <-res; err == nil {
				t.Log(n.Key, fmt.Sprintf("done (start: %v, end: %v, duration: %v)", start, time.Now(), time.Since(start)))
			} else {
				t.Log(n.Key, fmt.Sprintf("%v", err))
			}
			t.wg.Done()
		}(n, t)
	}
	t.wg.Wait()
	if t.logger != nil {
		t.logger.Wait()
	}
}

func (t *Thread) DryRun() {
	t.Prepare()
	for _, n := range t.Map {
		t.wg.Add(1)
		t.DryLog(n.Key, "pending...")
		go func(n *Node, t *Thread) {
			armed, res := n.DryRun()
			<-armed
			start := time.Now()
			t.DryLog(n.Key, fmt.Sprintf("executing... (start: %v)", start))
			if err := <-res; err == nil {
				t.DryLog(n.Key, fmt.Sprintf("done (start: %v, end: %v, duration: %v)", start, time.Now(), time.Since(start)))
			} else {
				t.DryLog(n.Key, fmt.Sprintf("failed (%v)", err))
			}
			t.wg.Done()
		}(n, t)
	}
	t.wg.Wait()
	if t.logger != nil {
		t.logger.Wait()
	}
}

func (t *Thread) RebuildToRegexp(pattern string) *Thread {
	if !t.Closed {
		t.Prepare()
	}
	nt := NewThread()
	nt.logger = t.logger
	for k := range t.Map {
		matched, err := regexp.MatchString(pattern, k)
		if err != nil {
			log.Fatal(err)
		}
		if matched {
			nt.RecBuildToNode(t.Map[k])
		}
	}
	nt.Prepare()
	return nt
}

func (t *Thread) RebuildToKey(k string) *Thread {
	if !t.Closed {
		t.Prepare()
	}
	nt := NewThread()
	nt.logger = t.logger
	nt.RecBuildToNode(t.Map[k])
	nt.Prepare()
	return nt
}

func (t *Thread) RecBuildToNode(n *Node) {
	t.Add(n.Key, n.KeyDeps, n.Function)
	for _, c := range n.Deps {
		t.RecBuildToNode(c)
	}
}

func (t *Thread) RebuildFromRegexp(pattern string) *Thread {
	if !t.Closed {
		t.Prepare()
	}
	nt := NewThread()
	nt.logger = t.logger
	for k := range t.Map {
		matched, err := regexp.MatchString(pattern, k)
		if err != nil {
			log.Fatal(err)
		}
		if matched {
			nt.RecBuildFromNode(t.Map, t.Map[k])
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
	nt.logger = t.logger
	nt.RecBuildFromNode(t.Map, t.Map[k])
	nt.Prepare()
	return nt
}

func (t *Thread) RecBuildFromNode(m MapNodes, n *Node) {
	t.Add(n.Key, n.KeyDeps, n.Function)
	for _, c := range m {
		if c.HasDep(n) {
			t.RecBuildFromNode(m, c)
		}
	}
}

func (t *Thread) ExcludeKeys(pattern string) *Thread {
	r := regexp.MustCompile(pattern)
	nt := NewThread()
	nt.logger = t.logger
	for key, node := range t.Map {
		if !r.MatchString(key) {
			nt.AddNode(node)
		}
	}
	return nt
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
