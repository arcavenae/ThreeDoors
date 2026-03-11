package dispatch

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// ErrQueueItemNotFound is returned when a queue item ID doesn't exist.
var ErrQueueItemNotFound = errors.New("queue item not found")

// DevQueue manages a persistent queue of dev dispatch items.
type DevQueue struct {
	mu    sync.RWMutex
	path  string
	items []QueueItem
	index map[string]int // id → index into items
}

// NewDevQueue creates a DevQueue backed by the given file path.
// The file and its parent directory are created if they don't exist.
func NewDevQueue(path string) (*DevQueue, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create queue directory: %w", err)
	}

	q := &DevQueue{
		path:  path,
		index: make(map[string]int),
	}

	if err := q.Load(path); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return q, nil
}

// Load reads queue items from the given YAML file.
func (q *DevQueue) Load(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	const maxConfigFileSize = 1 * 1024 * 1024 // 1MB
	if info.Size() > maxConfigFileSize {
		return fmt.Errorf("file %s exceeds size limit (%d bytes > %d bytes)", path, info.Size(), maxConfigFileSize)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	var items []QueueItem
	if err := yaml.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("parse dev queue: %w", err)
	}

	q.items = items
	q.rebuildIndex()
	return nil
}

// Save writes queue items to the given YAML file using atomic write.
func (q *DevQueue) Save(path string) error {
	data, err := yaml.Marshal(q.items)
	if err != nil {
		return fmt.Errorf("marshal dev queue: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create queue directory: %w", err)
	}

	return atomicWrite(path, data)
}

// Add inserts a new item into the queue. The item's ID is generated if empty.
func (q *DevQueue) Add(item QueueItem) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if item.ID == "" {
		item.ID = generateQueueID()
	}

	if _, exists := q.index[item.ID]; exists {
		return fmt.Errorf("queue item %s already exists", item.ID)
	}

	if item.Status == "" {
		item.Status = QueueItemPending
	}

	if item.QueuedAt == nil {
		now := time.Now().UTC()
		item.QueuedAt = &now
	}

	q.items = append(q.items, item)
	q.index[item.ID] = len(q.items) - 1

	return q.Save(q.path)
}

// Get returns the queue item with the given ID.
func (q *DevQueue) Get(id string) (QueueItem, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	idx, ok := q.index[id]
	if !ok {
		return QueueItem{}, fmt.Errorf("get queue item %s: %w", id, ErrQueueItemNotFound)
	}

	return q.items[idx], nil
}

// Update applies the given function to the queue item with the given ID and persists.
func (q *DevQueue) Update(id string, fn func(*QueueItem)) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	idx, ok := q.index[id]
	if !ok {
		return fmt.Errorf("update queue item %s: %w", id, ErrQueueItemNotFound)
	}

	fn(&q.items[idx])

	return q.Save(q.path)
}

// List returns a copy of all queue items.
func (q *DevQueue) List() []QueueItem {
	q.mu.RLock()
	defer q.mu.RUnlock()

	result := make([]QueueItem, len(q.items))
	copy(result, q.items)
	return result
}

// Remove deletes the queue item with the given ID from the queue and persists.
func (q *DevQueue) Remove(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	idx, ok := q.index[id]
	if !ok {
		return fmt.Errorf("remove queue item %s: %w", id, ErrQueueItemNotFound)
	}

	q.items = append(q.items[:idx], q.items[idx+1:]...)
	q.rebuildIndex()

	return q.Save(q.path)
}

// rebuildIndex reconstructs the id→index map from items.
func (q *DevQueue) rebuildIndex() {
	q.index = make(map[string]int, len(q.items))
	for i, item := range q.items {
		q.index[item.ID] = i
	}
}

// generateQueueID creates a queue item ID with "dq-" prefix + first 8 chars of UUID.
func generateQueueID() string {
	return "dq-" + uuid.New().String()[:8]
}

// atomicWrite writes data to path using the write-tmp/fsync/rename pattern.
func atomicWrite(path string, data []byte) error {
	tmpPath := path + ".tmp"

	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create temp file %s: %w", tmpPath, err)
	}

	writeOK := false
	defer func() {
		if !writeOK {
			_ = f.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write temp file %s: %w", tmpPath, err)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync temp file %s: %w", tmpPath, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("close temp file %s: %w", tmpPath, err)
	}

	writeOK = true

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename temp file to %s: %w", path, err)
	}

	return nil
}
