package store

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sync"
)

const (
	SlotSize     = 512
	HeaderSize   = 1 + 4 + 2 // state(1) + key(uint32) + len(uint16)
	PayloadCap   = SlotSize - HeaderSize
	StateEmpty   = 0
	StateOcc     = 1
	StateDeleted = 2
)

// Slot header is encoded as: state | key(uint32, LE) | payloadLen(uint16, LE)

var (
	ErrKeyNotFound   = errors.New("key not found")
	ErrTableFull     = errors.New("table full")
	ErrPayloadTooBig = errors.New("payload exceeds slot capacity")
)

type DB struct {
	f        *os.File
	slots    int
	modPrime uint32
	mu       sync.RWMutex
}

// Open creates or opens a fixed-slot DB file with given slot count.
// If the file does not exist, it initializes it to size slots*SlotSize with empty slots.
func Open(path string, slots int) (*DB, error) {
	if slots <= 0 {
		return nil, fmt.Errorf("slots must be > 0")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && !errors.Is(err, os.ErrExist) {
		return nil, err
	}

	// Open read-write, create if not exists
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}

	// Ensure file size matches number of slots
	wantSize := int64(slots * SlotSize)
	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if stat.Size() != wantSize {
		if err := f.Truncate(wantSize); err != nil {
			_ = f.Close()
			return nil, err
		}
		// Newly grown regions are zero-filled by the OS; zero state means empty.
	}

	db := &DB{
		f:        f,
		slots:    slots,
		modPrime: uint32(closestPrime(slots)),
	}
	return db, nil
}

func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.f == nil {
		return nil
	}
	err := db.f.Close()
	db.f = nil
	return err
}

// Stats represents counts of slot states in the DB file.
type Stats struct {
    Empty    int
    Occupied int
    Deleted  int
    Total    int
}

// Stats scans all slots and returns the distribution of states.
func (db *DB) Stats() (Stats, error) {
    db.mu.RLock()
    defer db.mu.RUnlock()
    var s Stats
    for i := 0; i < db.slots; i++ {
        state, _, _, err := db.readSlot(i)
        if err != nil {
            return s, err
        }
        s.Total++
        switch state {
        case StateEmpty:
            s.Empty++
        case StateOcc:
            s.Occupied++
        case StateDeleted:
            s.Deleted++
        }
    }
    return s, nil
}

// States returns a slice with the state byte of each slot in order.
func (db *DB) States() ([]byte, error) {
    db.mu.RLock()
    defer db.mu.RUnlock()
    out := make([]byte, db.slots)
    for i := 0; i < db.slots; i++ {
        state, _, _, err := db.readSlot(i)
        if err != nil {
            return nil, err
        }
        out[i] = state
    }
    return out, nil
}

// Clear resets all slots to StateEmpty and zero payloads.
func (db *DB) Clear() error {
    db.mu.Lock()
    defer db.mu.Unlock()
    zero := make([]byte, SlotSize)
    for i := 0; i < db.slots; i++ {
        off := int64(i * SlotSize)
        if _, err := db.f.WriteAt(zero, off); err != nil {
            return err
        }
    }
    return nil
}

var ErrKeyExists = errors.New("key already exists")

// Insert stores the value for the given string key.
// Fails with ErrKeyExists if the key is already present.
// Value is JSON-encoded with a small envelope that includes the original key and type name.
func (db *DB) Insert(key string, v any) error {
	payload, err := marshalEnvelope(key, v)
	if err != nil {
		return err
	}
	if len(payload) > PayloadCap {
		return fmt.Errorf("%w: %d > %d", ErrPayloadTooBig, len(payload), PayloadCap)
	}
	hk := hashKey(key)
	start := int(hk % db.modPrime)

	db.mu.Lock()
	defer db.mu.Unlock()

	// Linear probing: record first deleted slot to reuse if key not found
	firstDel := -1
	for probe := 0; probe < db.slots; probe++ {
		idx := (start + probe) % db.slots
		state, sk, existing, err := db.readSlot(idx)
		if err != nil {
			return err
		}
		switch state {
		case StateEmpty:
			if firstDel >= 0 {
				return db.writeSlot(firstDel, StateOcc, hk, payload)
			}
			return db.writeSlot(idx, StateOcc, hk, payload)
		case StateDeleted:
			if firstDel < 0 {
				firstDel = idx
			}
		case StateOcc:
			if sk == hk {
				// Verify actual key match to avoid hash collision overwriting
				env, derr := decodeEnvelope(existing)
				if derr == nil && env.Key == key {
					return ErrKeyExists
				}
			}
			// collision; continue probing
		default:
			// unknown state, treat as collision and continue
		}
	}
	if firstDel >= 0 {
		return db.writeSlot(firstDel, StateOcc, hk, payload)
	}
	return ErrTableFull
}

// Select loads the record for key into out. Returns (found=false) if not present.
func (db *DB) Select(key string, out any) (bool, error) {
	hk := hashKey(key)
	start := int(hk % db.modPrime)

	db.mu.RLock()
	defer db.mu.RUnlock()

	for probe := 0; probe < db.slots; probe++ {
		idx := (start + probe) % db.slots
		state, sk, payload, err := db.readSlot(idx)
		if err != nil {
			return false, err
		}
		switch state {
		case StateEmpty:
			// Empty slot terminates search in linear probing
			return false, nil
		case StateDeleted:
			// Keep probing
		case StateOcc:
			if sk == hk {
				env, derr := decodeEnvelope(payload)
				if derr == nil && env.Key == key {
					if err := json.Unmarshal(env.Data, out); err != nil {
						return false, err
					}
					return true, nil
				}
			}
		default:
			// continue
		}
	}
	return false, nil
}

// Delete removes the record for key if present. Returns (found=false) if it didn't exist.
func (db *DB) Delete(key string) (bool, error) {
	hk := hashKey(key)
	start := int(hk % db.modPrime)

	db.mu.Lock()
	defer db.mu.Unlock()

	for probe := 0; probe < db.slots; probe++ {
		idx := (start + probe) % db.slots
		state, sk, payload, err := db.readSlot(idx)
		if err != nil {
			return false, err
		}
		switch state {
		case StateEmpty:
			return false, nil
		case StateOcc:
			if sk == hk {
				env, derr := decodeEnvelope(payload)
				if derr == nil && env.Key == key {
					if err := db.writeSlot(idx, StateDeleted, 0, nil); err != nil {
						return false, err
					}
					return true, nil
				}
			}
		case StateDeleted:
			// continue
		}
	}
	return false, nil
}

// readSlot reads and decodes a slot at index into (state, key, payload)
func (db *DB) readSlot(index int) (byte, uint32, []byte, error) {
	off := int64(index * SlotSize)
	buf := make([]byte, SlotSize)
	if _, err := db.f.ReadAt(buf, off); err != nil {
		return 0, 0, nil, err
	}
	state := buf[0]
	key := binary.LittleEndian.Uint32(buf[1:5])
	plen := int(binary.LittleEndian.Uint16(buf[5:7]))
	if plen < 0 || plen > PayloadCap {
		return state, key, nil, fmt.Errorf("bad payload length")
	}
	payload := make([]byte, plen)
	copy(payload, buf[HeaderSize:HeaderSize+plen])
	return state, key, payload, nil
}

// writeSlot encodes and writes one slot at index
func (db *DB) writeSlot(index int, state byte, key uint32, payload []byte) error {
	if len(payload) > PayloadCap {
		return fmt.Errorf("%w: %d > %d", ErrPayloadTooBig, len(payload), PayloadCap)
	}
	buf := make([]byte, SlotSize)
	buf[0] = state
	binary.LittleEndian.PutUint32(buf[1:5], key)
	binary.LittleEndian.PutUint16(buf[5:7], uint16(len(payload)))
	copy(buf[HeaderSize:], payload)

	off := int64(index * SlotSize)
	_, err := db.f.WriteAt(buf, off)
	return err
}

// hashKey hashes a string key to uint32 using FNV-1a.
func hashKey(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

// closestPrime returns the prime number closest to n. If equidistant, returns the lower prime.
func closestPrime(n int) int {
	if n <= 2 {
		return 2
	}
	// quick check if n is prime
	if isPrime(n) {
		return n
	}
	lower := n - 1
	upper := n + 1
	for {
		lowerIsPrime := lower >= 2 && isPrime(lower)
		upperIsPrime := isPrime(upper)
		dl := math.MaxInt
		du := math.MaxInt
		if lowerIsPrime {
			dl = n - lower
		}
		if upperIsPrime {
			du = upper - n
		}
		if dl == du && dl != math.MaxInt {
			return lower
		}
		if dl < du {
			return lower
		}
		if du < dl {
			return upper
		}
		lower--
		upper++
	}
}

func isPrime(x int) bool {
	if x < 2 {
		return false
	}
	if x%2 == 0 {
		return x == 2
	}
	limit := int(math.Sqrt(float64(x)))
	for d := 3; d <= limit; d += 2 {
		if x%d == 0 {
			return false
		}
	}
	return true
}

// envelope stores the original key and JSON payload of the value.
type envelope struct {
	Key  string          `json:"key"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func marshalEnvelope(key string, v any) ([]byte, error) {
	// Determine a friendly type name
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	typeName := t.String()
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	env := envelope{Key: key, Type: typeName, Data: data}
	return json.Marshal(env)
}

func decodeEnvelope(b []byte) (*envelope, error) {
	var env envelope
	if err := json.Unmarshal(b, &env); err != nil {
		return nil, err
	}
	return &env, nil
}
