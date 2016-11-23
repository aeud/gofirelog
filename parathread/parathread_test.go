package parathread

import (
	// "fmt"
	"testing"
)

func TestStructure(t *testing.T) {
	th := NewThread()
	th.Add("test", []string{"foo", "bar"})
	th.Add("foo", []string{"bar"})
	th.Add("toto", []string{"foo"})
	th.Add("bar", []string{"adrien"})
	th.Add("bar2", []string{})
	th.Add("bar3", []string{})
	th.Prepare()
	// fmt.Println(th)
	th.Run(func(n *Node) {
		// fmt.Println(n.Key)
	})
}

func TestRebuild(t *testing.T) {
	th := NewThread()
	th.Add("test", []string{"foo", "bar"})
	th.Add("foo", []string{"bar"})
	th.Add("toto", []string{"foo"})
	th.Add("bar", []string{"adrien"})
	th.Add("bar2", []string{"foo"})
	th.Add("bar3", []string{})
	// fmt.Println(th.RebuildFromKey("test"))
	// fmt.Println(th.RebuildFromKey("foo"))
	// fmt.Println(th.RebuildFromKey("bar"))
	// fmt.Println(th.RebuildFromRegexp("bar.*"))

	th.RebuildFromRegexp("bar.*").Run(func(n *Node) {
		// fmt.Println(n.Key)
	})
}
