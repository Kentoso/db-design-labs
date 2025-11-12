package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Kentoso/db-design-labs/internal/store"
)

const defaultDBPath = "data/db.bin"

func usage() {
	exe := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s [ -db path ] select <key>\n", exe)
	fmt.Fprintf(os.Stderr, "  %s [ -db path ] insert <key> <json_payload>\n", exe)
	fmt.Fprintf(os.Stderr, "  %s [ -db path ] delete <key>\n", exe)
	fmt.Fprintf(os.Stderr, "  %s [ -db path ] scan\n", exe)
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
				fmt.Fprintln(os.Stderr, "key exists")
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
		fmt.Printf("empty %d\n", stats.Empty)
		fmt.Printf("occupied %d\n", stats.Occupied)
		fmt.Printf("deleted %d\n", stats.Deleted)
		fmt.Printf("total %d\n", stats.Total)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
	}
	return true
}
