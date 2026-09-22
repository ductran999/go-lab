// Command fetch downloads a URL in N parallel ranges and merges
// them into one file. Usage:
//
//	go run ./cmd/fetch -url http://localhost:8106/file -parts 5 -out ./tmp/lines.txt
package main

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	ErrNoRangeSupport   = errors.New("server does not accept ranges")
	ErrUnknownSize      = errors.New("unknown content size")
	ErrBadPart          = errors.New("part did not return 206")
	ErrBadURL           = errors.New("url must be http(s)")
	ErrChecksumMissing  = errors.New("no checksum endpoint")
	ErrChecksumMismatch = errors.New("merged file checksum mismatch")
)

func fail(err error) {
	fmt.Fprintln(os.Stderr, "fetch:", err)

	os.Exit(1)
}

// cleanURL parses the CLI-provided URL and allows http(s) only.
func cleanURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", ErrBadURL
	}

	return u.String(), nil
}

// probeSize asks what the server offers: ranges support, total size,
// and the Content-Digest checksum traveling with the response.
func probeSize(url string) (int64, string, error) {
	head, err := http.Head(url) //nolint // CLI downloader: URL restricted to http(s) by cleanURL.
	if err != nil {
		return 0, "", err
	}

	defer func() {
		_ = head.Body.Close()
	}()

	if head.Header.Get("Accept-Ranges") != "bytes" {
		return 0, "", ErrNoRangeSupport
	}

	size, err := strconv.ParseInt(head.Header.Get("Content-Length"), 10, 64)
	if err != nil || size <= 0 {
		return 0, "", ErrUnknownSize
	}

	digest := ""

	if cd := head.Header.Get("Content-Digest"); cd != "" {
		inner, _, _ := strings.Cut(strings.TrimPrefix(cd, "sha-256=:"), ":")
		digest = inner
	}

	if digest == "" {
		return 0, "", ErrChecksumMissing
	}

	return size, digest, nil
}

// setupFile prepares the output: dir exists, file created and sized.
func setupFile(path string, size int64) (*os.File, error) {
	mkErr := os.MkdirAll("./tmp", 0o750)
	if mkErr != nil {
		return nil, mkErr
	}

	//nolint // CLI downloader: output path is an explicit user flag.
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	truncErr := f.Truncate(size)
	if truncErr != nil {
		_ = f.Close()

		return nil, truncErr
	}

	return f, nil
}

// offsetWriter streams bytes to a fixed file offset. Memory stays
// flat (~32KB copy buffer) no matter how big the part is.
type offsetWriter struct {
	f   *os.File
	pos int64
}

func (w *offsetWriter) Write(p []byte) (int, error) {
	n, err := w.f.WriteAt(p, w.pos)
	w.pos += int64(n)

	return n, err
}

// verifyMerge hashes the merged file (streamed, flat memory) and
// compares against the digest the server published in Content-Digest.
// Ranges are not atomic — trust the checksum, never the download
// exit alone.
func verifyMerge(path, wantDigest string) error {
	//nolint // CLI downloader: verified file is the explicit user output.
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	hasher := sha256.New()

	_, err = io.Copy(hasher, f)
	if err != nil {
		return err
	}

	if base64.StdEncoding.EncodeToString(hasher.Sum(nil)) != wantDigest {
		return ErrChecksumMismatch
	}

	return nil
}
func fetchChunk(url string, start, end int64, f *os.File) error {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("%w: part %d-%d got %s", ErrBadPart, start, end, resp.Status)
	}

	written, err := io.Copy(&offsetWriter{f: f, pos: start}, resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("part bytes=%d-%d wrote %.2f MB\n", start, end, float64(written)/1048576)

	return nil
}

func main() {
	url := flag.String("url", "http://localhost:8106/file", "file URL (must accept ranges)")
	parts := flag.Int("parts", 5, "parallel ranges")
	out := flag.String("out", "./tmp/lines.txt", "merged output path")

	flag.Parse()

	target, err := cleanURL(*url)
	if err != nil {
		fail(err)
	}

	probe, digest, err := probeSize(target)
	if err != nil {
		fail(err)
	}

	fmt.Printf("probed %.2f MB, splitting into %d parts (~%.2f MB each)\n",
		float64(probe)/1048576, *parts, float64(probe)/float64(*parts)/1048576)

	f, err := setupFile(*out, probe)
	if err != nil {
		fail(err)
	}

	defer func() {
		_ = f.Close()
	}()

	var wg sync.WaitGroup

	errs := make(chan error, *parts)
	chunk := probe / int64(*parts)

	for i := range *parts {
		start := int64(i) * chunk
		end := probe - 1

		if i < *parts-1 {
			end = start + chunk - 1
		}

		wg.Go(func() {
			chunkErr := fetchChunk(target, start, end, f)
			if chunkErr != nil {
				errs <- chunkErr
			}
		})
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		fail(err)
	}

	fmt.Printf("merged %d bytes from %d parts into %s\n", probe, *parts, *out)

	verifyErr := verifyMerge(*out, digest)
	if verifyErr != nil {
		fail(verifyErr)
	}

	fmt.Println("checksum ok")
}
