// Package whatsapp wraps go.mau.fi/whatsmeow for clank's whatsapp subcommand.
//
// Persists the session in ~/.clank/whatsapp.db (modernc.org/sqlite, pure Go,
// no CGo). Exposes Open() to obtain an authenticated handle, Handle.Lookup
// for IsOnWhatsApp + UserInfo + ProfilePicture in one call, and Handle.Logout
// to unlink server-side and clear local state.
package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	_ "modernc.org/sqlite" // registers "sqlite" driver (pure Go, no CGo)
)

const (
	connectTimeout = 30 * time.Second
	dialect        = "sqlite"
)

func dbPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".clank")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "whatsapp.db"), nil
}

// dsn builds a modernc.org/sqlite DSN with foreign keys + WAL + busy timeout.
// Foreign keys are mandatory — sqlstore.Upgrade panics otherwise.
func dsn(path string) string {
	return "file:" + path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(wal)"
}

// Handle bundles a connected client and its container so we can clean both up.
type Handle struct {
	Client    *whatsmeow.Client
	container *sqlstore.Container

	loggedOut chan *events.LoggedOut
}

// Close disconnects the client and closes the underlying SQL container.
func (h *Handle) Close() {
	if h == nil {
		return
	}
	if h.Client != nil {
		h.Client.Disconnect()
	}
	if h.container != nil {
		_ = h.container.Close()
	}
}

// Open opens the SQLite store, runs QR pairing if no session is present, and
// blocks until events.Connected (auth completed). The caller MUST call Close.
//
// Set interactive=false to refuse to launch the QR flow when no session exists
// — useful for `lookup`, which should error rather than prompt for a scan
// when stdout might be a JSON pipe.
func Open(ctx context.Context, interactive bool) (*Handle, error) {
	path, err := dbPath()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}
	logger := waLog.Noop // use Stdout("clank-wa", "ERROR", false) to surface library errors

	container, err := sqlstore.New(ctx, dialect, dsn(path), logger)
	if err != nil {
		return nil, fmt.Errorf("open sqlstore at %s: %w", path, err)
	}
	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		_ = container.Close()
		return nil, fmt.Errorf("get device: %w", err)
	}

	client := whatsmeow.NewClient(device, logger)

	connected := make(chan struct{}, 1)
	loggedOut := make(chan *events.LoggedOut, 1)
	client.AddEventHandler(func(evt any) {
		switch e := evt.(type) {
		case *events.Connected:
			select {
			case connected <- struct{}{}:
			default:
			}
		case *events.LoggedOut:
			select {
			case loggedOut <- e:
			default:
			}
		}
	})

	h := &Handle{Client: client, container: container, loggedOut: loggedOut}

	if client.Store.ID == nil {
		if !interactive {
			h.Close()
			return nil, errors.New("not paired with WhatsApp — run `clank whatsapp login` first")
		}
		if err := pairWithQR(ctx, client); err != nil {
			h.Close()
			return nil, err
		}
	} else {
		if err := client.Connect(); err != nil {
			h.Close()
			return nil, fmt.Errorf("connect: %w", err)
		}
	}

	select {
	case <-connected:
		return h, nil
	case lo := <-loggedOut:
		h.Close()
		return nil, fmt.Errorf("logged out by server (reason=%v) — session cleared, run `clank whatsapp login` again", lo.Reason)
	case <-time.After(connectTimeout):
		h.Close()
		return nil, fmt.Errorf("timed out after %s waiting for events.Connected", connectTimeout)
	case <-ctx.Done():
		h.Close()
		return nil, ctx.Err()
	}
}

// pairWithQR drives the GetQRChannel flow and returns once the user has scanned
// (or an unrecoverable QR-channel event arrives).
func pairWithQR(ctx context.Context, client *whatsmeow.Client) error {
	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		return fmt.Errorf("qr channel: %w", err)
	}
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	fmt.Fprintln(os.Stderr, "WhatsApp pairing — open WhatsApp on your phone → Settings → Linked Devices → Link a Device.")
	for evt := range qrChan {
		switch evt.Event {
		case "code":
			fmt.Fprintln(os.Stderr)
			qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stderr)
			fmt.Fprintf(os.Stderr, "(scan within %s — a fresh code will replace this if you wait)\n", evt.Timeout)
		case "success":
			fmt.Fprintln(os.Stderr, "✓ paired successfully")
			// channel closes after this — loop exits naturally
		case "timeout":
			return errors.New("QR pairing timed out before scan")
		case "err-client-outdated":
			return errors.New("whatsmeow client is too old for current WhatsApp protocol — upgrade clank (`go install github.com/AnshumanAtrey/clank@latest`)")
		case "err-scanned-without-multidevice":
			fmt.Fprintln(os.Stderr, "Multi-device mode disabled in WhatsApp. Enable it in Settings → Linked Devices and re-scan.")
		case "err-unexpected-state":
			return errors.New("unexpected pairing state — try `clank whatsapp logout` then login again")
		case "error":
			return fmt.Errorf("pair error: %w", evt.Error)
		default:
			fmt.Fprintln(os.Stderr, "QR event:", evt.Event)
		}
	}
	return nil
}

// Result is the per-phone return shape from Lookup.
type Result struct {
	Query                string `json:"query"`
	JID                  string `json:"jid,omitempty"`
	LID                  string `json:"lid,omitempty"`
	Registered           bool   `json:"registered"`
	Reason               string `json:"reason,omitempty"`
	VerifiedBusinessName string `json:"verified_business_name,omitempty"`
	About                string `json:"about,omitempty"`
	PictureID            string `json:"picture_id,omitempty"`
	ProfilePictureURL    string `json:"profile_picture_url,omitempty"`
	DeviceCount          int    `json:"device_count,omitempty"`
}

// Lookup runs IsOnWhatsApp + GetUserInfo + GetProfilePictureInfo for one phone.
// Phone should be E.164 with leading + (e.g. "+14155552671"). The library
// silently drops numbers it considers invalid (issue #1086) — when that
// happens we synthesise a Reason instead of returning a misleading "registered=false".
func (h *Handle) Lookup(ctx context.Context, phone string) (*Result, error) {
	if h == nil || h.Client == nil {
		return nil, errors.New("nil handle — open the client first")
	}
	resps, err := h.Client.IsOnWhatsApp(ctx, []string{phone})
	if err != nil {
		return nil, fmt.Errorf("IsOnWhatsApp: %w", err)
	}

	out := &Result{Query: phone}

	// Reconcile by Query, not by index — see issue #1086.
	var hit *types.IsOnWhatsAppResponse
	stripped := stripPlus(phone)
	for i := range resps {
		r := &resps[i]
		if r.Query == phone || r.Query == stripped || stripPlus(r.Query) == stripped {
			hit = r
			break
		}
	}
	if hit == nil {
		out.Reason = "no response from WhatsApp — likely invalid format (whatsmeow #1086)"
		return out, nil
	}

	out.JID = hit.JID.String()
	out.Registered = hit.IsIn
	if hit.VerifiedName != nil && hit.VerifiedName.Details != nil {
		out.VerifiedBusinessName = hit.VerifiedName.Details.GetVerifiedName()
	}
	if !hit.IsIn {
		return out, nil
	}

	infoMap, err := h.Client.GetUserInfo(ctx, []types.JID{hit.JID})
	if err == nil {
		if info, ok := infoMap[hit.JID]; ok {
			out.About = info.Status
			out.PictureID = info.PictureID
			out.DeviceCount = len(info.Devices)
			if !info.LID.IsEmpty() {
				out.LID = info.LID.String()
			}
		}
	}

	pic, err := h.Client.GetProfilePictureInfo(ctx, hit.JID, &whatsmeow.GetProfilePictureParams{Preview: false})
	if err == nil && pic != nil {
		out.ProfilePictureURL = pic.URL
	}
	return out, nil
}

// Logout unlinks server-side via the remove-companion-device IQ, then deletes
// the local SQLite file. Returns the path that was deleted (for logging).
func (h *Handle) Logout(ctx context.Context) (string, error) {
	if h == nil || h.Client == nil {
		return "", errors.New("nil handle")
	}
	logoutErr := h.Client.Logout(ctx)
	h.Client.Disconnect()
	_ = h.container.Close()

	path, pathErr := dbPath()
	if pathErr == nil {
		_ = os.Remove(path)
		_ = os.Remove(path + "-wal")
		_ = os.Remove(path + "-shm")
	}
	return path, logoutErr
}

// Reset deletes the local SQLite file without contacting WhatsApp. Useful when
// the session is already invalid and Logout would just hang on socket dial.
func Reset() (string, error) {
	path, err := dbPath()
	if err != nil {
		return "", err
	}
	_ = os.Remove(path)
	_ = os.Remove(path + "-wal")
	_ = os.Remove(path + "-shm")
	return path, nil
}

func stripPlus(s string) string {
	if len(s) > 0 && s[0] == '+' {
		return s[1:]
	}
	return s
}
