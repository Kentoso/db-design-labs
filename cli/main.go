package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Kentoso/db-design-labs/internal/store"
)

const defaultDBPath = "data/db.bin"

func usage() {
	exe := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s [ -db path ] select <key>\n", exe)
	fmt.Fprintf(os.Stderr, "  %s [ -db path ] insert <key> <json_payload>\n", exe)
	fmt.Fprintf(os.Stderr, "  %s [ -db path ] delete <key>\n", exe)
}

func main() {
	dbPath := flag.String("db", defaultDBPath, "database file path")
	flag.Parse()

	if flag.NArg() < 2 {
		usage()
		os.Exit(2)
	}

	cmd := flag.Arg(0)
	key := flag.Arg(1)

	// Always open or create with 5000 slots as requested
	db, err := store.Open(*dbPath, 5000)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	switch cmd {
	case "select":
		var raw json.RawMessage
		found, err := db.Select(key, &raw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "select: %v\n", err)
			os.Exit(1)
		}
		if !found {
			fmt.Fprintln(os.Stderr, "not found")
			os.Exit(1)
		}
		if len(raw) == 0 {
			fmt.Println("null")
			return
		}
		fmt.Println(string(raw))

	case "insert":
		if flag.NArg() < 3 {
			fmt.Fprintln(os.Stderr, "insert requires <key> <json_payload>")
			os.Exit(2)
		}
		payload := flag.Arg(2)
		// Validate JSON
		var tmp any
		if err := json.Unmarshal([]byte(payload), &tmp); err != nil {
			fmt.Fprintf(os.Stderr, "invalid json payload: %v\n", err)
			os.Exit(2)
		}
		// Store raw JSON to preserve shape
		var raw json.RawMessage = json.RawMessage(payload)
		if err := db.Insert(key, &raw); err != nil {
			if errors.Is(err, store.ErrKeyExists) {
				fmt.Fprintln(os.Stderr, "key exists")
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "insert: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("ok")

	case "delete":
		deleted, err := db.Delete(key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "delete: %v\n", err)
			os.Exit(1)
		}
		if !deleted {
			fmt.Fprintln(os.Stderr, "not found")
			os.Exit(1)
		}
		fmt.Println("ok")

	default:
		usage()
		os.Exit(2)
	}
}
