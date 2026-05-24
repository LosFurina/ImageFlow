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

// imageRef represents a found local image reference in markdown
type imageRef struct {
	rawRef  string // original text, e.g. "![[xxx.jpg]]" or "![alt](path)"
	altText string // alt text for replacement
	absPath string // absolute file path on disk
	isWiki  bool   // true if Obsidian wiki link ![[...]]
}

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

// FindLocalImages scans markdown for local image references.
// Supports both standard markdown ![alt](path) and Obsidian wiki links ![[filename]].
// For wiki links, resolves against vaultRoot/assets/.
func FindLocalImages(mdContent string, mdDir string, vaultRoot string) map[string]imageRef {
	result := make(map[string]imageRef)

	// 1. Obsidian wiki links: ![[filename]] or ![[path/to/file.png]]
	wikiRe := regexp.MustCompile(`!\[\[([^\]]+)\]\]`)
	for _, m := range wikiRe.FindAllStringSubmatch(mdContent, -1) {
		if len(m) < 2 {
			continue
		}
		rawRef := m[0]
		filename := strings.TrimSpace(m[1])

		// For wiki links, try vaultRoot/assets/filename first
		absPath := filepath.Join(vaultRoot, "assets", filename)
		if _, err := os.Stat(absPath); err != nil {
			// Fallback: try relative to markdown file dir
			absPath = filepath.Join(mdDir, filename)
			if _, err := os.Stat(absPath); err != nil {
				continue
			}
		}

		altText := filename
		if idx := strings.LastIndex(altText, "."); idx > 0 {
			altText = altText[:idx]
		}

		result[rawRef] = imageRef{
			rawRef:  rawRef,
			altText: altText,
			absPath: absPath,
			isWiki:  true,
		}
	}

	// 2. Standard markdown links: ![alt](path)
	mdRe := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	for _, m := range mdRe.FindAllStringSubmatch(mdContent, -1) {
		if len(m) < 3 {
			continue
		}
		rawRef := m[0]
		altText := m[1]
		ref := strings.TrimSpace(m[2])

		// Skip URLs and data URIs
		if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") || strings.HasPrefix(ref, "data:") {
			continue
		}

		// Resolve to absolute path
		absPath := ref
		if !filepath.IsAbs(ref) {
			absPath = filepath.Join(mdDir, ref)
		}

		if _, err := os.Stat(absPath); err == nil {
			result[rawRef] = imageRef{
				rawRef:  rawRef,
				altText: altText,
				absPath: absPath,
				isWiki:  false,
			}
		}
	}

	return result
}

// moveToAssets moves the original image file to vaultRoot/assets/<articleName>/
func moveToAssets(absPath string, vaultRoot string, articleName string) (string, error) {
	targetDir := filepath.Join(vaultRoot, "assets", articleName)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("create assets dir: %w", err)
	}

	baseName := filepath.Base(absPath)
	targetPath := filepath.Join(targetDir, baseName)

	// Handle duplicate filenames
	if _, err := os.Stat(targetPath); err == nil {
		ext := filepath.Ext(baseName)
		name := strings.TrimSuffix(baseName, ext)
		targetPath = filepath.Join(targetDir, fmt.Sprintf("%s_%d%s", name, time.Now().Unix(), ext))
	}

	if err := os.Rename(absPath, targetPath); err != nil {
		return "", fmt.Errorf("move file: %w", err)
	}
	return targetPath, nil
}

func main() {
	var (
		endpoint = flag.String("endpoint", defaultEndpoint, "ImageFlow server endpoint")
		inPlace  = flag.Bool("i", false, "Edit file in place")
		output   = flag.String("o", "", "Output file path (default: stdout)")
		vault    = flag.String("vault", "", "Obsidian vault root directory (for resolving ![[...]] wiki links). Defaults to markdown file's parent dir.")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <markdown-file>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Upload local images in a Markdown file to ImageFlow and replace references.\n")
		fmt.Fprintf(os.Stderr, "Supports standard Markdown ![alt](path) and Obsidian wiki links ![[filename]].\n")
		fmt.Fprintf(os.Stderr, "AK/SK credentials MUST be set via environment variables (never pass as flags).\n\n")
		fmt.Fprintf(os.Stderr, "Required environment variables:\n")
		fmt.Fprintf(os.Stderr, "  IMAGEFLOW_AK                 Access Key\n")
		fmt.Fprintf(os.Stderr, "  IMAGEFLOW_SK                 Secret Key\n")
		fmt.Fprintf(os.Stderr, "  IMAGEFLOW_OPENAPI_ENDPOINT   ImageFlow OpenAPI endpoint (optional, overrides -endpoint flag)\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Standard markdown\n")
		fmt.Fprintf(os.Stderr, "  export IMAGEFLOW_AK=xxx IMAGEFLOW_SK=yyy\n")
		fmt.Fprintf(os.Stderr, "  %s article.md\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Obsidian vault with wiki links\n")
		fmt.Fprintf(os.Stderr, "  export IMAGEFLOW_AK=xxx IMAGEFLOW_SK=yyy\n")
		fmt.Fprintf(os.Stderr, "  %s -vault ~/obsidian/note article.md\n", os.Args[0])
	}
	flag.Parse()

	// Environment variable overrides flag for endpoint
	if envEndpoint := os.Getenv("IMAGEFLOW_OPENAPI_ENDPOINT"); envEndpoint != "" {
		*endpoint = envEndpoint
	}

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

	vaultRoot := *vault
	if vaultRoot == "" {
		vaultRoot = mdDir
	}

	// Article name (without .md extension) for asset organization
	articleName := strings.TrimSuffix(filepath.Base(mdPath), filepath.Ext(mdPath))

	content, err := os.ReadFile(mdPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}
	mdContent := string(content)

	// Find local images
	images := FindLocalImages(mdContent, mdDir, vaultRoot)
	if len(images) == 0 {
		fmt.Fprintln(os.Stderr, "No local images found.")
		fmt.Print(mdContent)
		return
	}

	fmt.Fprintf(os.Stderr, "Found %d local image(s):\n", len(images))
	for _, ref := range images {
		fmt.Fprintf(os.Stderr, "  - %s\n", ref.rawRef)
	}

	// Upload each image and collect replacements
	replacements := make(map[string]string) // rawRef -> new markdown
	movedFiles := make(map[string]bool)     // dedup: track already-moved files
	for _, ref := range images {
		fmt.Fprintf(os.Stderr, "Uploading %s ... ", filepath.Base(ref.absPath))
		url, err := UploadImage(*endpoint, ak, sk, ref.absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAILED: %v\n", err)
			continue
		}
		fmt.Fprintf(os.Stderr, "OK -> %s\n", url)

		// Build replacement markdown
		newRef := fmt.Sprintf("![%s](%s)", ref.altText, url)
		replacements[ref.rawRef] = newRef

		// Move original file to assets/<articleName>/
		if !movedFiles[ref.absPath] {
			newPath, err := moveToAssets(ref.absPath, vaultRoot, articleName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  WARN: failed to move original file: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "  Moved original to %s\n", newPath)
				movedFiles[ref.absPath] = true
			}
		}
	}

	if len(replacements) == 0 {
		fmt.Fprintln(os.Stderr, "No images uploaded successfully.")
		fmt.Print(mdContent)
		os.Exit(1)
	}

	// Replace references in markdown
	newContent := mdContent
	for rawRef, newRef := range replacements {
		newContent = strings.ReplaceAll(newContent, rawRef, newRef)
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
