package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "errors"
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "strings"

	"github.com/Kentoso/db-design-labs/internal/store"
)

const defaultDBPath = "data/db.bin"

// run represents a contiguous occupied region in the slot array
type run struct{ start, length int }

func usage() {
	exe := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s [ -db path ] select <key>\n", exe)
	fmt.Fprintf(os.Stderr, "  %s [ -db path ] insert <key> <json_payload>\n", exe)
	fmt.Fprintf(os.Stderr, "  %s [ -db path ] delete <key>\n", exe)
	fmt.Fprintf(os.Stderr, "  %s [ -db path ] scan [threshold]\n", exe)
	fmt.Fprintf(os.Stderr, "  %s [ -db path ] clear\n", exe)
}

func main() {
	dbPath := flag.String("db", defaultDBPath, "database file path")
	flag.Parse()

	db, err := store.Open(*dbPath, 5000)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// If args provided, process once, then continue reading lines (REPL or piped)
	if flag.NArg() > 0 {
		line := strings.Join(flag.Args(), " ")
		_ = processLine(db, line)
	}

	// Terminal detection for prompt
	fi, _ := os.Stdin.Stat()
	interactive := fi.Mode()&os.ModeCharDevice != 0

	sc := bufio.NewScanner(os.Stdin)
	for {
		if interactive {
			fmt.Print("db> ")
		}
		if !sc.Scan() {
			break
		}
		line := strings.TrimSpace(sc.Text())
		if cont := processLine(db, line); !cont {
			break
		}
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "read: %v\n", err)
		os.Exit(1)
	}
}

// processLine executes a single line. Returns false to exit loop.
func processLine(db *store.DB, line string) bool {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return true
	}
	switch line {
	case "exit", "quit", "q":
		return false
	case "help":
		usage()
		return true
	}
	parts := strings.SplitN(line, " ", 3)
	cmd := parts[0]
	switch cmd {
	case "insert":
		if len(parts) < 3 {
			fmt.Fprintln(os.Stderr, "insert requires <key> <json_payload>")
			return true
		}
		key := parts[1]
		payload := parts[2]
		var tmp any
		if err := json.Unmarshal([]byte(payload), &tmp); err != nil {
			fmt.Fprintf(os.Stderr, "invalid json payload for key %s: %v\n", key, err)
			return true
		}
		var raw json.RawMessage = json.RawMessage(payload)
		if err := db.Insert(key, &raw); err != nil {
			if errors.Is(err, store.ErrKeyExists) {
				fmt.Fprintln(os.Stderr, fmt.Sprintf("key %s exists", key))
				return true
			}
			fmt.Fprintf(os.Stderr, "insert: %v\n", err)
			return true
		}
		fmt.Println("ok")
	case "select":
		if len(parts) < 2 {
			fmt.Fprintln(os.Stderr, "select requires <key>")
			return true
		}
		key := parts[1]
		var raw json.RawMessage
		found, err := db.Select(key, &raw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "select: %v\n", err)
			return true
		}
		if !found {
			fmt.Fprintln(os.Stderr, "not found")
			return true
		}
		if len(raw) == 0 {
			fmt.Println("null")
			return true
		}
		fmt.Println(string(raw))
	case "delete":
		if len(parts) < 2 {
			fmt.Fprintln(os.Stderr, "delete requires <key>")
			return true
		}
		key := parts[1]
		deleted, err := db.Delete(key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "delete: %v\n", err)
			return true
		}
		if !deleted {
			fmt.Fprintln(os.Stderr, "not found")
			return true
		}
		fmt.Println("ok")
	case "scan":
		stats, err := db.Stats()
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan: %v\n", err)
			return true
		}
		states, err := db.States()
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan states: %v\n", err)
			return true
		}
		// optional threshold argument
		threshold := 1
		if len(parts) >= 2 {
			if v, err := strconv.Atoi(parts[1]); err == nil && v > 0 {
				threshold = v
			}
		}
		lf := 0.0
		if stats.Total > 0 {
			lf = float64(stats.Occupied) / float64(stats.Total)
		}
		fmt.Printf("empty %d\n", stats.Empty)
		fmt.Printf("occupied %d\n", stats.Occupied)
		fmt.Printf("deleted %d\n", stats.Deleted)
		fmt.Printf("total %d\n", stats.Total)
		fmt.Printf("load_factor %.4f\n", lf)
        // Dense zones: contiguous occupied runs; report top 10 and persist all filtered
        var runs []run
		for i := 0; i < len(states); {
			if states[i] != store.StateOcc {
				i++
				continue
			}
			j := i
			for j < len(states) && states[j] == store.StateOcc {
				j++
			}
			runs = append(runs, run{start: i, length: j - i})
			i = j
		}
		// sort by length desc (selection sort)
		for a := 0; a < len(runs); a++ {
			maxIdx := a
			for b := a + 1; b < len(runs); b++ {
				if runs[b].length > runs[maxIdx].length {
					maxIdx = b
				}
			}
			runs[a], runs[maxIdx] = runs[maxIdx], runs[a]
		}
		// filter by threshold
		filtered := make([]run, 0, len(runs))
		for _, r := range runs {
			if r.length >= threshold {
				filtered = append(filtered, r)
			}
		}
		fmt.Printf("dense_zones %d (threshold %d)\n", len(filtered), threshold)
        // write all details to dense_zones.txt (in current directory)
        if err := writeDenseZonesFile("dense_zones.txt", db, filtered); err != nil {
            fmt.Fprintf(os.Stderr, "write dense_zones.txt: %v\n", err)
        } else {
            fmt.Println("saved dense_zones.txt")
        }
		// show top 10 on stdout
		limit := 10
		if len(filtered) < limit {
			limit = len(filtered)
		}
		for i := 0; i < limit; i++ {
			fmt.Printf("zone %d start %d length %d\n", i+1, filtered[i].start, filtered[i].length)
		}
	case "clear":
		if err := db.Clear(); err != nil {
			fmt.Fprintf(os.Stderr, "clear: %v\n", err)
			return true
		}
		fmt.Println("ok")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
	}
	return true
}

func writeDenseZonesFile(path string, db *store.DB, runs []run) error {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()
    // header
    if _, err := fmt.Fprintf(f, "# Dense zones report\n# Each section shows contiguous occupied slots with deserialized payloads.\n\n"); err != nil {
        return err
    }
    sep := strings.Repeat("=", 72)
    for zi, r := range runs {
        if _, err := fmt.Fprintf(f, "%s\nZONE %d  start %d  length %d\n%s\n", sep, zi+1, r.start, r.length, sep); err != nil {
            return err
        }
        for pos := r.start; pos < r.start+r.length; pos++ {
            d, err := db.SlotDetail(pos)
            if err != nil {
                return err
            }
            // pretty-print JSON data
            pretty := d.Data
            if len(d.Data) > 0 {
                var buf bytes.Buffer
                if err := json.Indent(&buf, d.Data, "", "  "); err == nil {
                    pretty = buf.Bytes()
                }
            }
            if _, err := fmt.Fprintf(f, "- position: %d\n  key: %s\n  hash: %d\n  data: %s\n\n", d.Index, d.Key, d.Hash, string(pretty)); err != nil {
                return err
            }
        }
    }
    return nil
}
