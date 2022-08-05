package bookstack

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"go.uber.org/ratelimit"
)

const (
	appJSON = "application/json"
)

type Bookstack struct {
	url         string
	tokenID     string
	tokenSecret string
	limit       ratelimit.Limiter
	log         *log.Logger
	insecure    bool
}

type Option func(*Bookstack)

func SetLogger(l *log.Logger) Option {
	return func(b *Bookstack) {
		b.log = l
	}
}

func SetToken(id, secret string) Option {
	return func(b *Bookstack) {
		b.tokenID = id
		b.tokenSecret = secret
	}
}

// SetURL sets the url of the site to control.
func SetURL(url string) Option {
	return func(b *Bookstack) {
		b.url = url
	}
}

func SetRateLimit(limit int) Option {
	return func(b *Bookstack) {
		b.limit = ratelimit.New(limit)
	}
}

func New(opts ...Option) *Bookstack {

	b := &Bookstack{
		limit: ratelimit.New(180),
		log:   log.New(ioutil.Discard, "", 0),
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

func (b *Bookstack) authorization() string {
	return fmt.Sprintf("Token %s:%s", b.tokenID, b.tokenSecret)
}

type Form interface {
	Form() (string, io.Reader, error)
}

func (b *Bookstack) request(ctx context.Context, method, query string, data Form) ([]byte, error) {

	b.limit.Take()

	url := fmt.Sprintf("%s/api/%s", strings.TrimRight(b.url, "/"), strings.TrimLeft(query, "/"))

	client := http.DefaultClient

	if b.insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	contentType, reader, err := data.Form()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", b.authorization())
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode <= http.StatusIMUsed {
		return raw, nil
	}

	msg := Response{}

	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, err
	}

	return nil, msg.Error()

}

type Single interface {
	User | Book | BookDetailed | Chapter | ChapterDetailed | Page | PageDetailed | Shelf | ShelfDetailed | RecycledBook | RecycledPage | RecycledChapter | Attachment | AttachmentDetailed
}

type Group interface {
	[]User | []Book | []Chapter | []Page | []Shelf | []RecycleBinItem | []Attachment
}

func ParseSingle[s Single](data []byte) (s, error) {

	var result s

	if err := json.Unmarshal(data, &result); err != nil {
		return result, nil
	}

	return result, nil
}

func ParseMultiple[g Group](data []byte) (g, error) {

	r := Response{}

	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}

	if r.Error() != nil {
		return nil, r.Error()
	}

	var result g

	if err := json.Unmarshal(r.Data, &result); err != nil {
		return nil, err
	}

	return result, nil
}
