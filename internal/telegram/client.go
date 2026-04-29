package telegram

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"golang.org/x/term"
)

const (
	envAppID   = "TG_APP_ID"
	envAppHash = "TG_APP_HASH"
)

type Lookup struct {
	Phone      string `json:"phone"`
	Found      bool   `json:"found"`
	UserID     int64  `json:"user_id,omitempty"`
	FirstName  string `json:"first_name,omitempty"`
	LastName   string `json:"last_name,omitempty"`
	Username   string `json:"username,omitempty"`
	Premium    bool   `json:"premium,omitempty"`
	Verified   bool   `json:"verified,omitempty"`
	Bot        bool   `json:"bot,omitempty"`
	Restricted bool   `json:"restricted,omitempty"`
	HasPhoto   bool   `json:"has_photo,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

func sessionPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".clank")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "telegram.session"), nil
}

func credentials() (int, string, error) {
	idStr := os.Getenv(envAppID)
	hash := os.Getenv(envAppHash)
	if idStr == "" || hash == "" {
		return 0, "", fmt.Errorf("set %s and %s (get them from https://my.telegram.org/apps)", envAppID, envAppHash)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, "", fmt.Errorf("invalid %s: %w", envAppID, err)
	}
	return id, hash, nil
}

func newClient() (*telegram.Client, error) {
	appID, appHash, err := credentials()
	if err != nil {
		return nil, err
	}
	sp, err := sessionPath()
	if err != nil {
		return nil, err
	}
	return telegram.NewClient(appID, appHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: sp},
	}), nil
}

func Login(ctx context.Context) error {
	client, err := newClient()
	if err != nil {
		return err
	}
	return client.Run(ctx, func(ctx context.Context) error {
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return err
		}
		if status.Authorized {
			fmt.Println("already logged in as", describeUser(status.User))
			return nil
		}

		phone := promptLine("Phone (E.164, e.g. +14155552671): ")
		flow := auth.NewFlow(
			authenticator{phone: phone},
			auth.SendCodeOptions{},
		)
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
		status2, _ := client.Auth().Status(ctx)
		if status2.Authorized {
			fmt.Println("logged in as", describeUser(status2.User))
		}
		return nil
	})
}

func Logout(ctx context.Context) error {
	sp, err := sessionPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(sp); errors.Is(err, os.ErrNotExist) {
		fmt.Println("no session to clear")
		return nil
	}
	if err := os.Remove(sp); err != nil {
		return err
	}
	fmt.Println("removed", sp)
	return nil
}

func ResolvePhone(ctx context.Context, phone string) (Lookup, error) {
	out := Lookup{Phone: phone}
	digits := strings.TrimPrefix(strings.TrimSpace(phone), "+")
	if digits == "" {
		return out, errors.New("empty phone")
	}

	client, err := newClient()
	if err != nil {
		return out, err
	}

	err = client.Run(ctx, func(ctx context.Context) error {
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return err
		}
		if !status.Authorized {
			return errors.New("not logged in — run `clank telegram login` first")
		}
		api := client.API()
		res, err := api.ContactsResolvePhone(ctx, digits)
		if err != nil {
			if tgerr.Is(err, "PHONE_NOT_OCCUPIED") {
				out.Found = false
				out.Reason = "phone not registered (or hidden by privacy)"
				return nil
			}
			if tgerr.Is(err, "PHONE_NUMBER_INVALID") {
				out.Reason = "phone number invalid"
				return err
			}
			if e, ok := tgerr.As(err); ok && strings.HasPrefix(e.Type, "FLOOD_WAIT") {
				out.Reason = fmt.Sprintf("rate-limited — wait %d seconds", e.Argument)
				return err
			}
			return err
		}
		for _, u := range res.Users {
			if user, ok := u.(*tg.User); ok {
				out.Found = true
				out.UserID = user.ID
				out.FirstName = user.FirstName
				out.LastName = user.LastName
				out.Username = user.Username
				out.Premium = user.Premium
				out.Verified = user.Verified
				out.Bot = user.Bot
				out.Restricted = user.Restricted
				_, hasPhoto := user.GetPhoto()
				out.HasPhoto = hasPhoto
				return nil
			}
		}
		out.Reason = "resolved peer was not a User"
		return nil
	})
	return out, err
}

type authenticator struct {
	phone string
}

func (a authenticator) Phone(_ context.Context) (string, error) {
	if a.phone != "" {
		return a.phone, nil
	}
	return promptLine("Phone (E.164): "), nil
}

func (a authenticator) Password(_ context.Context) (string, error) {
	fmt.Print("2FA password (cloud password — empty if not set): ")
	pw, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(pw), nil
}

func (a authenticator) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	return promptLine("Login code (sent via SMS / Telegram app): "), nil
}

func (a authenticator) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
	fmt.Println("Telegram TOS:", tos.Text)
	return nil
}

func (a authenticator) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("clank does not register new accounts; sign up via the official Telegram app first")
}

func promptLine(prompt string) string {
	fmt.Print(prompt)
	r := bufio.NewReader(os.Stdin)
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

func describeUser(u tg.UserClass) string {
	if u == nil {
		return "(unknown)"
	}
	if user, ok := u.(*tg.User); ok {
		parts := []string{}
		if user.FirstName != "" {
			parts = append(parts, user.FirstName)
		}
		if user.LastName != "" {
			parts = append(parts, user.LastName)
		}
		s := strings.Join(parts, " ")
		if user.Username != "" {
			s += " (@" + user.Username + ")"
		}
		s += fmt.Sprintf(" [id=%d]", user.ID)
		return s
	}
	return "(unknown user class)"
}
