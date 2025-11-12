package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "math/rand"
    "os"
    "path/filepath"
    "strings"
    "time"
)

var (
    outPath = flag.String("out", "generated.txt", "output file path")
    count   = flag.Int("n", 100, "number of lines to generate")
    model   = flag.String("model", "client", "model name to generate (client, employee, campaign, ad_platform, campaign_platform, ad_set, media_asset, video, image, ad_text, ad)")
    seed    = flag.Int64("seed", time.Now().UnixNano(), "random seed")
)

func main() {
    flag.Parse()
    rand.Seed(*seed)

    if *count <= 0 {
        fmt.Fprintln(os.Stderr, "n must be > 0")
        os.Exit(2)
    }

    if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil && !os.IsExist(err) {
        fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
        os.Exit(1)
    }
    f, err := os.Create(*outPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "create: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()

    m := strings.ToLower(*model)
    for i := 1; i <= *count; i++ {
        key, payload := generate(m, i)
        line := fmt.Sprintf("insert %s %s\n", key, string(payload))
        if _, err := f.WriteString(line); err != nil {
            fmt.Fprintf(os.Stderr, "write: %v\n", err)
            os.Exit(1)
        }
    }
}

func generate(model string, id int) (string, []byte) {
    ts := time.Now().UTC().Format(time.RFC3339)
    switch model {
    case "client":
        key := fmt.Sprintf("client:%d", id)
        obj := map[string]any{
            "id":        id,
            "name":      fmt.Sprintf("Client %d", id),
            "email":     fmt.Sprintf("client%03d@example.com", id),
            "createdAt": ts,
        }
        return key, mustJSON(obj)
    case "employee":
        key := fmt.Sprintf("employee:%d", id)
        obj := map[string]any{
            "id":        id,
            "name":      fmt.Sprintf("Employee %d", id),
            "position":  pick([]string{"Engineer", "Manager", "Designer"}, id),
            "createdAt": ts,
        }
        return key, mustJSON(obj)
    case "campaign":
        key := fmt.Sprintf("campaign:%d", id)
        start := time.Now().UTC().AddDate(0, 0, rand.Intn(10))
        finish := start.AddDate(0, 0, 7+rand.Intn(14))
        obj := map[string]any{
            "id":         id,
            "name":       fmt.Sprintf("Campaign %d", id),
            "startDate":  start.Format("2006-01-02T15:04:05Z07:00"),
            "finishDate": finish.Format("2006-01-02T15:04:05Z07:00"),
            "clientId":   1 + (id % 5),
            "managerId":  1 + (id % 3),
            "createdAt":  ts,
        }
        return key, mustJSON(obj)
    case "ad_platform":
        key := fmt.Sprintf("ad_platform:%d", id)
        obj := map[string]any{"id": id, "name": pick([]string{"Meta", "Google", "TikTok", "X"}, id)}
        return key, mustJSON(obj)
    case "campaign_platform":
        key := fmt.Sprintf("campaign_platform:%d", id)
        obj := map[string]any{
            "campaignId": 1 + (id % 7),
            "platformId": 1 + (id % 4),
            "budgetCents": 10000 + (id % 5000),
        }
        return key, mustJSON(obj)
    case "ad_set":
        key := fmt.Sprintf("ad_set:%d", id)
        obj := map[string]any{
            "id":           id,
            "name":         fmt.Sprintf("AdSet %d", id),
            "campaignId":   1 + (id % 7),
            "targetAge":    "18-35",
            "targetGender": "any",
            "targetCountry": "US",
            "createdAt":    ts,
        }
        return key, mustJSON(obj)
    case "media_asset":
        key := fmt.Sprintf("media_asset:%d", id)
        obj := map[string]any{
            "id":           id,
            "name":         fmt.Sprintf("asset_%d", id),
            "filePath":     fmt.Sprintf("/assets/%d.bin", id),
            "creationDate": ts,
        }
        return key, mustJSON(obj)
    case "video":
        key := fmt.Sprintf("video:%d", id)
        obj := map[string]any{"mediaAssetId": id, "duration": 30 + (id % 60)}
        return key, mustJSON(obj)
    case "image":
        key := fmt.Sprintf("image:%d", id)
        obj := map[string]any{"mediaAssetId": id, "resolution": "1080x1080"}
        return key, mustJSON(obj)
    case "ad_text":
        key := fmt.Sprintf("ad_text:%d", id)
        obj := map[string]any{"id": id, "text": fmt.Sprintf("Buy now %d!", id), "createdAt": ts}
        return key, mustJSON(obj)
    case "ad":
        key := fmt.Sprintf("ad:%d", id)
        obj := map[string]any{
            "id":           id,
            "adSetId":      1 + (id % 7),
            "mediaAssetId": 1 + (id % 11),
            "adTextId":     1 + (id % 13),
            "createdAt":    ts,
        }
        return key, mustJSON(obj)
    default:
        // fallback to client
        key := fmt.Sprintf("client:%d", id)
        obj := map[string]any{
            "id":        id,
            "name":      fmt.Sprintf("Client %d", id),
            "email":     fmt.Sprintf("client%03d@example.com", id),
            "createdAt": ts,
        }
        return key, mustJSON(obj)
    }
}

func mustJSON(v any) []byte {
    b, _ := json.Marshal(v)
    return b
}

func pick[T any](arr []T, idx int) T {
    return arr[idx%len(arr)]
}
