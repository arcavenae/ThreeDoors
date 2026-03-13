package device

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestLocalDeviceRegistry_Register(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	reg := NewLocalDeviceRegistry(filepath.Join(dir, "devices"))

	id, err := NewDeviceID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewDeviceID() unexpected error: %v", err)
	}

	dev := Device{
		ID:        id,
		Name:      "test-laptop",
		FirstSeen: time.Now().UTC(),
		LastSync:  time.Now().UTC(),
	}

	if err := reg.Register(dev); err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}
}

func TestLocalDeviceRegistry_RegisterDuplicate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	reg := NewLocalDeviceRegistry(filepath.Join(dir, "devices"))

	id, err := NewDeviceID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewDeviceID() unexpected error: %v", err)
	}

	dev := Device{
		ID:        id,
		Name:      "test-laptop",
		FirstSeen: time.Now().UTC(),
		LastSync:  time.Now().UTC(),
	}

	if err := reg.Register(dev); err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	err = reg.Register(dev)
	if !errors.Is(err, ErrDeviceAlreadyExists) {
		t.Errorf("Register() duplicate = %v, want ErrDeviceAlreadyExists", err)
	}
}

func TestLocalDeviceRegistry_Get(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	reg := NewLocalDeviceRegistry(filepath.Join(dir, "devices"))

	id, err := NewDeviceID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewDeviceID() unexpected error: %v", err)
	}

	dev := Device{
		ID:        id,
		Name:      "test-laptop",
		FirstSeen: time.Now().UTC(),
		LastSync:  time.Now().UTC(),
	}

	if err := reg.Register(dev); err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	got, err := reg.Get(id)
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}

	if got.ID != id {
		t.Errorf("Get() ID = %s, want %s", got.ID, id)
	}
	if got.Name != "test-laptop" {
		t.Errorf("Get() Name = %s, want test-laptop", got.Name)
	}
}

func TestLocalDeviceRegistry_GetNotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	reg := NewLocalDeviceRegistry(filepath.Join(dir, "devices"))

	id, err := NewDeviceID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewDeviceID() unexpected error: %v", err)
	}

	_, err = reg.Get(id)
	if !errors.Is(err, ErrDeviceNotFound) {
		t.Errorf("Get() non-existent = %v, want ErrDeviceNotFound", err)
	}
}

func TestLocalDeviceRegistry_ListEmpty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	reg := NewLocalDeviceRegistry(filepath.Join(dir, "devices"))

	devices, err := reg.List()
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}

	if devices == nil {
		t.Error("List() returned nil, want empty slice")
	}
	if len(devices) != 0 {
		t.Errorf("List() returned %d devices, want 0", len(devices))
	}
}

func TestLocalDeviceRegistry_ListMultiple(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	reg := NewLocalDeviceRegistry(filepath.Join(dir, "devices"))

	ids := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"660e8400-e29b-41d4-a716-446655440001",
	}

	for i, rawID := range ids {
		id, err := NewDeviceID(rawID)
		if err != nil {
			t.Fatalf("NewDeviceID(%s) unexpected error: %v", rawID, err)
		}
		dev := Device{
			ID:        id,
			Name:      "device-" + string(rune('a'+i)),
			FirstSeen: time.Now().UTC(),
			LastSync:  time.Now().UTC(),
		}
		if err := reg.Register(dev); err != nil {
			t.Fatalf("Register() unexpected error: %v", err)
		}
	}

	devices, err := reg.List()
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}

	if len(devices) != 2 {
		t.Errorf("List() returned %d devices, want 2", len(devices))
	}
}

func TestLocalDeviceRegistry_Update(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	reg := NewLocalDeviceRegistry(filepath.Join(dir, "devices"))

	id, err := NewDeviceID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewDeviceID() unexpected error: %v", err)
	}

	dev := Device{
		ID:        id,
		Name:      "old-name",
		FirstSeen: time.Now().UTC(),
		LastSync:  time.Now().UTC(),
	}

	if err := reg.Register(dev); err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	dev.Name = "new-name"
	if err := reg.Update(dev); err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}

	got, err := reg.Get(id)
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}

	if got.Name != "new-name" {
		t.Errorf("Get() Name = %s, want new-name", got.Name)
	}
}

func TestLocalDeviceRegistry_Remove(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	reg := NewLocalDeviceRegistry(filepath.Join(dir, "devices"))

	id, err := NewDeviceID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewDeviceID() unexpected error: %v", err)
	}

	dev := Device{
		ID:        id,
		Name:      "test-laptop",
		FirstSeen: time.Now().UTC(),
		LastSync:  time.Now().UTC(),
	}

	if err := reg.Register(dev); err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	if err := reg.Remove(id); err != nil {
		t.Fatalf("Remove() unexpected error: %v", err)
	}

	_, err = reg.Get(id)
	if !errors.Is(err, ErrDeviceNotFound) {
		t.Errorf("Get() after Remove = %v, want ErrDeviceNotFound", err)
	}
}

func TestLocalDeviceRegistry_RemoveNotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	reg := NewLocalDeviceRegistry(filepath.Join(dir, "devices"))

	id, err := NewDeviceID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewDeviceID() unexpected error: %v", err)
	}

	err = reg.Remove(id)
	if !errors.Is(err, ErrDeviceNotFound) {
		t.Errorf("Remove() non-existent = %v, want ErrDeviceNotFound", err)
	}
}

func TestLocalDeviceRegistry_AutoCreatesDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	regDir := filepath.Join(dir, "nested", "devices")
	reg := NewLocalDeviceRegistry(regDir)

	id, err := NewDeviceID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("NewDeviceID() unexpected error: %v", err)
	}

	dev := Device{
		ID:        id,
		Name:      "test",
		FirstSeen: time.Now().UTC(),
		LastSync:  time.Now().UTC(),
	}

	if err := reg.Register(dev); err != nil {
		t.Fatalf("Register() should auto-create directory, got: %v", err)
	}
}

func TestLocalDeviceRegistry_ConcurrentRegister(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	reg := NewLocalDeviceRegistry(filepath.Join(dir, "devices"))

	var wg sync.WaitGroup
	errs := make([]error, 10)

	for i := range 10 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// Each goroutine registers a unique device
			rawID := "550e8400-e29b-41d4-a716-44665544" + string(rune('0'+idx/10)) + string(rune('0'+idx%10)) + "00"
			id, err := NewDeviceID(rawID)
			if err != nil {
				errs[idx] = err
				return
			}
			dev := Device{
				ID:        id,
				Name:      "concurrent-device",
				FirstSeen: time.Now().UTC(),
				LastSync:  time.Now().UTC(),
			}
			errs[idx] = reg.Register(dev)
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: Register() error: %v", i, err)
		}
	}
}
