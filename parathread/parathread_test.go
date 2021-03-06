package parathread

import (
	"fmt"
	"os"
	"testing"
	"time"
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
	if false {
		endpoint := os.Getenv("FIRELOG_ENDPOINT")
		authToken := os.Getenv("FIRELOG_AUTHTOKEN")
		ref := time.Now().Format("2006-01-02T15:04:05")
		th.AddLogger(endpoint, authToken, ref)
	}
	th.Run(func(n *Node) error {
		fmt.Println(n.Key)
		return nil
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
	newTh := th.RebuildFromRegexp("bar.*")
	newTh.Run(func(n *Node) error {
		// fmt.Println(n.Key)
		return nil
	})
}

func TestRebuildDesc(t *testing.T) {
	th := NewThread()
	th.Add("test", []string{"foo", "bar"})
	th.Add("foo", []string{"bar"})
	th.Add("toto", []string{"foo"})
	th.Add("bar", []string{"adrien"})
	th.Add("bar2", []string{"foo"})
	th.Add("bar3", []string{})
	fmt.Println(th.RebuildFromKey("foo"))
	fmt.Println(th.RebuildToKey("foo"))
	fmt.Println(th.RebuildFromRegexp("bar"))
	fmt.Println(th.RebuildFromRegexp("test"))
}
