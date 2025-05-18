package global

import "sync"

var (
	masterToUpdate string
	mu             sync.Mutex
)

// SetMasterToUpdate sets the special bypass key
func SetMasterToUpdate(value string) {
	mu.Lock()
	defer mu.Unlock()
	masterToUpdate = value
}

// GetMasterToUpdate retrieves the current bypass key
func GetMasterToUpdate() string {
	mu.Lock()
	defer mu.Unlock()
	return masterToUpdate
}

// ResetMasterToUpdate clears the bypass key (optional)
func ResetMasterToUpdate() {
	mu.Lock()
	defer mu.Unlock()
	masterToUpdate = ""
}
