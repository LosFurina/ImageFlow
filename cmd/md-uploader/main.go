package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	defaultEndpoint = "http://localhost:8686"
)

// SignRequest computes the AK/SK HMAC signature for a request
func SignRequest(ak, sk, method, path string, body []byte) (signature, timestamp string) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	bodyHash := sha256Body(body)
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s", method, path, ts, bodyHash)
	mac := hmac.New(sha256.New, []byte(sk))
	mac.Write([]byte(stringToSign))
	sig := hex.EncodeToString(mac.Sum(nil))
	return sig, ts
}

func sha256Body(body []byte) string {
	if len(body) == 0 {
		body = []byte("")
	}
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

// UploadImage uploads a single image file to ImageFlow via OpenAPI
func UploadImage(endpoint, ak, sk string, imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("open image: %w", err)
	}
	defer file.Close()

	// Build multipart body first (so we can sign it)
	var bodyBuf bytes.Buffer
	writer := multipart.NewWriter(&bodyBuf)
	part, err := writer.CreateFormFile("images[]", filepath.Base(imagePath))
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("copy file: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close writer: %w", err)
	}

	bodyBytes := bodyBuf.Bytes()
	path := "/openapi/upload"
	sig, ts := SignRequest(ak, sk, "POST", path, bodyBytes)

	req, err := http.NewRequest("POST", endpoint+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Access-Key", ak)
	req.Header.Set("X-Signature", sig)
	req.Header.Set("X-Timestamp", ts)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Results []struct {
			URLs struct {
				Original string `json:"original"`
				WebP     string `json:"webp"`
				AVIF     string `json:"avif"`
			} `json:"urls"`
		} `json:"results"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	if len(result.Results) == 0 {
		return "", fmt.Errorf("no image returned")
	}
	// Prefer original URL, fallback to webp
	url := result.Results[0].URLs.Original
	if url == "" {
		url = result.Results[0].URLs.WebP
	}
	if url == "" {
		return "", fmt.Errorf("no URL in response")
	}
	// Make absolute URL if relative
	if !strings.HasPrefix(url, "http") {
		url = endpoint + url
	}
	return url, nil
}

// FindLocalImages scans markdown for local image references
// Returns a map of original reference -> absolute image path
func FindLocalImages(mdContent string, mdDir string) map[string]string {
	// Match ![alt](path) where path does not start with http:// or https://
	re := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	matches := re.FindAllStringSubmatch(mdContent, -1)

	result := make(map[string]string)
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		ref := m[2] // the path inside parentheses
		ref = strings.TrimSpace(ref)

		// Skip URLs
		if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
			continue
		}
		// Skip data URIs
		if strings.HasPrefix(ref, "data:") {
			continue
		}

		// Resolve to absolute path
		absPath := ref
		if !filepath.IsAbs(ref) {
			absPath = filepath.Join(mdDir, ref)
		}

		// Only include if file exists
		if _, err := os.Stat(absPath); err == nil {
			result[ref] = absPath
		}
	}
	return result
}

type replacement struct {
	oldRef string
	newURL string
}

func main() {
	var (
		endpoint = flag.String("endpoint", defaultEndpoint, "ImageFlow server endpoint")
		inPlace  = flag.Bool("i", false, "Edit file in place")
		output   = flag.String("o", "", "Output file path (default: stdout)")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <markdown-file>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Upload local images in a Markdown file to ImageFlow and replace references.\n")
		fmt.Fprintf(os.Stderr, "AK/SK credentials MUST be set via environment variables (never pass as flags).\n\n")
		fmt.Fprintf(os.Stderr, "Required environment variables:\n")
		fmt.Fprintf(os.Stderr, "  IMAGEFLOW_AK    Access Key\n")
		fmt.Fprintf(os.Stderr, "  IMAGEFLOW_SK    Secret Key\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  export IMAGEFLOW_AK=xxx IMAGEFLOW_SK=yyy\n")
		fmt.Fprintf(os.Stderr, "  %s article.md\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  export IMAGEFLOW_AK=xxx IMAGEFLOW_SK=yyy\n")
		fmt.Fprintf(os.Stderr, "  %s -i article.md\n", os.Args[0])
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	ak := os.Getenv("IMAGEFLOW_AK")
	sk := os.Getenv("IMAGEFLOW_SK")
	if ak == "" || sk == "" {
		fmt.Fprintln(os.Stderr, "Error: IMAGEFLOW_AK and IMAGEFLOW_SK environment variables are required.")
		fmt.Fprintln(os.Stderr, "Never pass AK/SK as command-line flags (visible in shell history and process list).")
		os.Exit(1)
	}

	mdPath := flag.Arg(0)
	mdDir := filepath.Dir(mdPath)

	content, err := os.ReadFile(mdPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}
	mdContent := string(content)

	// Find local images
	images := FindLocalImages(mdContent, mdDir)
	if len(images) == 0 {
		fmt.Fprintln(os.Stderr, "No local images found.")
		fmt.Print(mdContent)
		return
	}

	fmt.Fprintf(os.Stderr, "Found %d local image(s):\n", len(images))
	for ref := range images {
		fmt.Fprintf(os.Stderr, "  - %s\n", ref)
	}

	// Upload each image and collect replacements
	var reps []replacement
	for ref, absPath := range images {
		fmt.Fprintf(os.Stderr, "Uploading %s ... ", ref)
		url, err := UploadImage(*endpoint, ak, sk, absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAILED: %v\n", err)
			continue
		}
		fmt.Fprintf(os.Stderr, "OK -> %s\n", url)
		reps = append(reps, replacement{oldRef: ref, newURL: url})
	}

	if len(reps) == 0 {
		fmt.Fprintln(os.Stderr, "No images uploaded successfully.")
		fmt.Print(mdContent)
		os.Exit(1)
	}

	// Replace references in markdown
	newContent := mdContent
	for _, r := range reps {
		// Escape the old ref for regex (it might contain special chars like dots)
		escaped := regexp.QuoteMeta(r.oldRef)
		// Replace ![alt](oldRef) with ![alt](newURL)
		re := regexp.MustCompile(`(!\[[^\]]*\]\()` + escaped + `\)`)
		newContent = re.ReplaceAllString(newContent, "${1}"+r.newURL+")")
	}

	// Output
	if *inPlace {
		if err := os.WriteFile(mdPath, []byte(newContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Updated %s in place.\n", mdPath)
	} else if *output != "" {
		if err := os.WriteFile(*output, []byte(newContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Written to %s\n", *output)
	} else {
		fmt.Print(newContent)
	}
}
