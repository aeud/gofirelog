package firelog

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestFirebase(t *testing.T) {
	ref := time.Now().Format("2006-01-02T15:04:05")
	endpoint := os.Getenv("FIRELOG_ENDPOINT")
	authToken := os.Getenv("FIRELOG_AUTHTOKEN")
	logs := NewFirebaseService(endpoint, authToken, ref)
	logs.Deamon()
	logs.Push(NewFirebaseMessage("k", "c"))
	logs.Push(NewFirebaseMessage("coucou", "c"))
	logs.Push(NewFirebaseMessage("coucou", "d"))
	logs.Push(NewFirebaseMessage("coucou", "e"))
	logs.Push(NewFirebaseMessage("coucou", "f"))
	logs.Push(NewFirebaseMessage("coucou", "g"))
	logs.Push(NewFirebaseMessage("coucou", "h"))
	logs.Push(NewFirebaseMessage("coucou", "i"))
	logs.Wait()
}
