package main

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/3mro777o/qari/internal/store"
)

const releasesURL = "https://github.com/3mro777o/qari/releases/latest"

// checkUpdate looks for a newer release in the background; the channel gets
// the notice to show, or "".
func checkUpdate(c *http.Client, st *store.Store) <-chan string {
	ch := make(chan string, 1)
	go func(u store.UpdateCheck) {
		u, s := updateNotice(u, version, time.Now(), func() (string, error) { return latestRelease(c) })
		st.SetRelease(u)
		ch <- s
	}(st.Release)
	return ch
}

// updateNotice asks latest at most once a day, and returns a notice when a
// newer version is out, at most once a day for the same version. Builds
// that are not a release (dev, snapshots) never check.
func updateNotice(u store.UpdateCheck, current string, now time.Time, latest func() (string, error)) (store.UpdateCheck, string) {
	if _, ok := semver(current); !ok {
		return u, ""
	}
	if now.Sub(u.Checked) >= 24*time.Hour {
		u.Checked = now
		if v, err := latest(); err == nil {
			u.Latest = v
		}
	}
	if !newer(u.Latest, current) || (u.Shown == u.Latest && now.Sub(u.ShownAt) < 24*time.Hour) {
		return u, ""
	}
	u.Shown, u.ShownAt = u.Latest, now
	return u, fmt.Sprintf("qari %s is out · run the install command again to update", strings.TrimPrefix(u.Latest, "v"))
}

// latestRelease is the tag GitHub redirects releases/latest to.
func latestRelease(c *http.Client) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, releasesURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	tag := path.Base(resp.Request.URL.Path)
	if _, ok := semver(tag); !ok {
		return "", fmt.Errorf("no release tag in %s", resp.Request.URL)
	}
	return tag, nil
}

// semver parses "1.2.3" or "v1.2.3"; anything else (dev, pre-releases,
// snapshots) is not a release.
func semver(v string) ([]int, bool) {
	parts := strings.Split(strings.TrimPrefix(v, "v"), ".")
	if len(parts) != 3 {
		return nil, false
	}
	out := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return nil, false
		}
		out[i] = n
	}
	return out, true
}

func newer(latest, current string) bool {
	l, ok1 := semver(latest)
	c, ok2 := semver(current)
	return ok1 && ok2 && slices.Compare(l, c) > 0
}
