package firelog

import (
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
	logs.Wait()
}
