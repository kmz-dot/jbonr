// Package creator runs the signup pipeline for one account:
// disposable inbox -> Jango signup -> email verification -> session capture.
package creator

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"codeberg.org/0xgetz/jango-bulk/internal/jango"
	"codeberg.org/0xgetz/jango-bulk/internal/mailtm"
	"codeberg.org/0xgetz/jango-bulk/internal/store"
)

// Options configures the creator.
type Options struct {
	Concurrency      int                                                                          // parallel signups (default 4)
	NamePrefix       string                                                                       // display name prefix (default "Jango User")
	Name             string                                                                       // fixed display name; overrides NamePrefix when set
	EmailPrefix      string                                                                       // mailbox local-part prefix (default "jango")
	MailPassword     string                                                                       // mailbox password (default random)
	AccountPassword  string                                                                       // Jango password (default "Str0ng!Pass123")
	MailAPI          string                                                                       // mail.tm base (default https://api.mail.tm)
	JangoAPI         string                                                                       // Jango API base
	VerifyTimeout    time.Duration                                                                // mail wait (default 180s)
	SkipVerification bool                                                                         // still captures session, just skips the link
	OnProgress       func(Event)                                                                  // optional progress hook
	SignupFunc       func(ctx context.Context, name, email, password string) (*jango.User, error) // test seam
}

// Event is a progress update.
type Event struct {
	Email string `json:"email"`
	Stage string `json:"stage"`
	Ok    bool   `json:"ok"`
	Err   string `json:"err,omitempty"`
}

func (o *Options) defaults() {
	if o.Concurrency <= 0 {
		o.Concurrency = 4
	}
	if o.NamePrefix == "" {
		o.NamePrefix = "Jango User"
	}
	if o.EmailPrefix == "" {
		o.EmailPrefix = "jango"
	}
	if o.AccountPassword == "" {
		o.AccountPassword = "Str0ng!Pass123"
	}
	if o.VerifyTimeout <= 0 {
		o.VerifyTimeout = 180 * time.Second
	}
}

// Creator runs batches.
type Creator struct {
	Opts Options
}

func New(o Options) *Creator {
	o.defaults()
	return &Creator{Opts: o}
}

const passAlphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#$%"

func randToken(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = passAlphabet[rand.Intn(len(passAlphabet))]
	}
	return string(b)
}

var counter int64

// Run creates n accounts, writing each result to st as soon as it is ready.
// Errors are per-account: one failure never aborts the batch.
func (c *Creator) Run(ctx context.Context, st *store.Store, n int, report func(store.Account)) {
	jobs := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < c.Opts.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				acc := c.one(ctx)
				if acc != nil {
					if err := st.Upsert(acc); err != nil {
						acc.Error = "store: " + err.Error()
					}
					_ = st.Log(map[bool]string{true: "created", false: "failed"}[acc.Error == ""],
						map[string]any{"email": acc.Email, "id": acc.ID, "credits": acc.Credits,
							"error": acc.Error})
					report(*acc)
				}
			}
		}()
	}
	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			i = n
		case jobs <- struct{}{}:
		}
	}
	close(jobs)
	wg.Wait()
}

func (c *Creator) emit(e Event) {
	if c.Opts.OnProgress != nil {
		c.Opts.OnProgress(e)
	}
}

// Signup is a single account (exported for tests and the `signup` command).
func (c *Creator) Signup(ctx context.Context) *store.Account { return c.one(ctx) }

func (c *Creator) one(ctx context.Context) *store.Account {
	ctx, cancel := context.WithTimeout(ctx, c.Opts.VerifyTimeout+120*time.Second)
	defer cancel()
	n := atomic.AddInt64(&counter, 1)
	name := c.Opts.Name
	if name == "" {
		name = fmt.Sprintf("%s %d", c.Opts.NamePrefix, n)
	}
	mailPass := c.Opts.MailPassword
	if mailPass == "" {
		mailPass = randToken(18)
	}
	acc := &store.Account{
		Name:          name,
		Password:      c.Opts.AccountPassword,
		InboxPassword: mailPass,
		CreatedAt:     time.Now().UTC(),
	}
	fail := func(stage string, err error) *store.Account {
		acc.Error = stage + ": " + err.Error()
		c.emit(Event{Email: acc.Email, Stage: stage, Err: err.Error()})
		return acc
	}

	mail := mailtm.New(c.Opts.MailAPI)
	ib, err := mail.Create(ctx, mailtm.RandomLocal(c.Opts.EmailPrefix, 12), mailPass)
	if err != nil {
		return fail("inbox", err)
	}
	acc.Email = ib.Address
	c.emit(Event{Email: acc.Email, Stage: "inbox", Ok: true})

	jc := jango.New(c.Opts.JangoAPI)
	if ib.Token == "" {
		_ = mail.Login(ctx, ib)
	}
	signup := jc.Signup
	if c.Opts.SignupFunc != nil {
		signup = c.Opts.SignupFunc
	}
	if _, err := signup(ctx, name, ib.Address, acc.Password); err != nil && !jango.IsStatus(err, 409) {
		return fail("signup", err)
	}
	c.emit(Event{Email: acc.Email, Stage: "signup", Ok: true})

	if !c.Opts.SkipVerification {
		link, _, err := mail.WaitLink(ctx, ib, func(u string) bool {
			return strings.Contains(u, "verifyEmail") || strings.Contains(u, "/verify")
		}, c.Opts.VerifyTimeout, 4*time.Second)
		if err != nil {
			return fail("verify-mail", err)
		}
		if err := jc.VerifyEmail(ctx, link, baseOrigin(link)); err != nil {
			return fail("verify", err)
		}
		acc.Verified = true
		c.emit(Event{Email: acc.Email, Stage: "verified", Ok: true})
	}

	// make sure we have a fresh access token, then capture both credentials
	if _, err := jc.Token(ctx); err != nil {
		return fail("session", err)
	}
	me, err := jc.Me(ctx)
	if err != nil {
		return fail("profile", err)
	}
	acc.ID = me.ID
	acc.FirebaseUID = me.FirebaseUID
	acc.AccessToken = jc.AccessToken()
	acc.TokenExpiry = jc.AccessExpiry()
	acc.RefreshCookie = jc.RefreshCookie()
	if acc.RefreshCookie == "" {
		return fail("session", fmt.Errorf("no %s cookie captured", jango.RefreshCookie))
	}
	acc.CookieExpiry = time.Now().UTC().Add(30 * 24 * time.Hour)
	if me.Credits != "" {
		fmt.Sscan(me.Credits, &acc.Credits)
	}
	acc.UpdatedAt = time.Now().UTC()
	c.emit(Event{Email: acc.Email, Stage: "session", Ok: true})
	return acc
}

// baseOrigin returns scheme://host of a URL, so the Bearer fetch is sent to the API
// origin while the cookie jar still receives the action page's cookies.
func baseOrigin(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}
