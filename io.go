package span

import (
	"archive/zip"
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

// LinkReader implements io.Reader for a URL.
type LinkReader struct {
	Link string
	buf  bytes.Buffer
	once sync.Once
}

// fill copies the content of the URL into the internal buffer.
func (r *LinkReader) fill() (err error) {
	r.once.Do(func() {
		var resp *http.Response
		resp, err = http.Get(r.Link)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		_, err = io.Copy(&r.buf, resp.Body)
	})
	return err
}

func (r *LinkReader) Read(p []byte) (int, error) {
	if err := r.fill(); err != nil {
		return 0, err
	}
	return r.buf.Read(p)
}

// SavedLink saves the content of a URL to a file.
type SavedLink struct {
	Link string
	f    *os.File
}

// Save link to a temporary file, return the filename.
func (s *SavedLink) Save() (filename string, err error) {
	r := &LinkReader{Link: s.Link}
	s.f, err = ioutil.TempFile("", "span-")
	if err != nil {
		return "", err
	}
	defer s.f.Close()
	if _, err := io.Copy(s.f, r); err != nil {
		return "", err
	}
	return s.f.Name(), nil
}

// Remove remove any left over temporary file.
func (s *SavedLink) Remove() {
	_ = os.Remove(s.f.Name())
}

// ZipContentReader returns all files in zip concatenated.
type ZipContentReader struct {
	Filename string
	buf      bytes.Buffer
	once     sync.Once
}

// fill populates the internal buffer with the content of all archive members.
func (r *ZipContentReader) fill() (err error) {
	r.once.Do(func() {
		var rc *zip.ReadCloser
		if rc, err = zip.OpenReader(r.Filename); err != nil {
			return
		}
		defer rc.Close()

		for _, f := range rc.File {
			var frc io.ReadCloser
			if frc, err = f.Open(); err != nil {
				return
			}
			if _, err = io.Copy(&r.buf, frc); err != nil {
				return
			}
			err = frc.Close()
		}
	})
	return
}

// Read returns the content of all archive members.
func (r *ZipContentReader) Read(p []byte) (int, error) {
	if err := r.fill(); err != nil {
		return 0, err
	}
	return r.buf.Read(p)
}

// FileReader creates a ReadCloser from a filename. If postpones error handling up
// until the first read.
type FileReader struct {
	Filename string
	f        *os.File
}

// Read reads from the file.
func (r *FileReader) Read(p []byte) (n int, err error) {
	if r.f == nil {
		r.f, err = os.Open(r.Filename)
		if err != nil {
			return
		}
	}
	n, err = r.f.Read(p)
	if err == io.EOF {
		defer r.f.Close()
		r.f = nil
	}
	return
}

// Close closes the file.
func (r *FileReader) Close() error {
	if r.f != nil {
		return r.f.Close()
	}
	return nil
}

// ZipOrPlainLinkReader is a reader that transparently handles zipped and uncompressed
// content. If the content is zipped, all archive entries are extracted.
type ZipOrPlainLinkReader struct {
	Link string
	buf  bytes.Buffer
	once sync.Once
}

// fill fills the internal buffer.
func (r *ZipOrPlainLinkReader) fill() (err error) {
	r.once.Do(func() {
		var filename string
		link := SavedLink{Link: r.Link}
		filename, err = link.Save()
		if err != nil {
			return
		}
		defer link.Remove()

		zipReader := &ZipContentReader{Filename: filename}
		// If there is no error with zip, assume it was a zip and return.
		if _, err = io.Copy(&r.buf, zipReader); err == nil {
			return
		}
		// Error with zip? Return plain content.
		_, err = io.Copy(&r.buf, &FileReader{Filename: filename})
	})
	return err
}

// Read implements the reader interface.
func (r *ZipOrPlainLinkReader) Read(p []byte) (int, error) {
	if err := r.fill(); err != nil {
		return 0, err
	}
	return r.buf.Read(p)
}

// ZipOrPlainLinkMultiReader reads uncompressed or zipped content from multiple URLs.
type ZipOrPlainLinkMultiReader struct {
	Links []string
	r     io.Reader
}

// Read reads content from all links.
func (r *ZipOrPlainLinkMultiReader) Read(p []byte) (int, error) {
	if r.r == nil {
		var readers []io.Reader
		for _, link := range r.Links {
			readers = append(readers, &ZipOrPlainLinkReader{Link: link})
		}
		r.r = io.MultiReader(readers...)
	}
	return r.r.Read(p)
}

// SavedReaders takes a list of readers and persists their content in temporary file.
type SavedReaders struct {
	Readers []io.Reader
	f       *os.File
}

// Save saves all readers to a temporary file and returns the filename.
func (r *SavedReaders) Save() (filename string, err error) {
	r.f, err = ioutil.TempFile("", "span-")
	if err != nil {
		return
	}
	if _, err = io.Copy(r.f, io.MultiReader(r.Readers...)); err != nil {
		return
	}
	if err = r.f.Close(); err != nil {
		return
	}
	filename = r.f.Name()
	return
}

// Remove remove any left over temporary file.
func (r *SavedReaders) Remove() {
	_ = os.Remove(r.f.Name())
}

// ReadLines returns a list of trimmed lines in a file. Empty lines are skipped.
func ReadLines(filename string) (lines []string, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return
}

// WriteCounter counts the number of bytes written through it.
type WriteCounter struct {
	w     io.Writer
	count uint64
}

func (w *WriteCounter) Write(p []byte) (int, error) {
	atomic.AddUint64(&w.count, uint64(len(p)))
	return len(p), nil
}

// Count returns the number of bytes written.
func (w *WriteCounter) Count() uint64 {
	return atomic.LoadUint64(&w.count)
}
