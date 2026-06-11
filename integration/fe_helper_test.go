//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tebeka/selenium"
)

const geckoPort = 8080

var geckoDriverPath string

func init() {
	if p := os.Getenv("GECKODRIVER_PATH"); p != "" {
		geckoDriverPath = p
	} else {
		_, f, _, _ := runtime.Caller(0)
		geckoDriverPath = filepath.Join(filepath.Dir(f), "vendor", "geckodriver")
	}
}

func getRemote(t *testing.T) selenium.WebDriver {
	t.Helper()

	fbuf, err := selenium.NewFrameBuffer()
	require.NoError(t, err)
	t.Cleanup(func() {
		fbuf.Stop()
	})

	cmd := exec.Command(geckoDriverPath, "--port", strconv.Itoa(geckoPort))
	cmd.Env = append(os.Environ(), "DISPLAY=:"+fbuf.Display, "XAUTHORITY="+fbuf.AuthPath)
	err = cmd.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		cmd.Process.Kill()
		cmd.Wait()
	})

	addr := fmt.Sprintf("http://127.0.0.1:%d", geckoPort)
	for i := 0; i < 30; i++ {
		time.Sleep(time.Second)
		resp, err := http.Get(addr + "/status")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
	}

	caps := selenium.Capabilities{"browserName": "firefox"}
	wd, err := selenium.NewRemote(caps, addr)
	require.NoError(t, err)
	t.Cleanup(func() {
		wd.Quit()
	})

	return wd
}

type waitForFn func(*testing.T, selenium.WebDriver) bool

func waitFor(t *testing.T, wd selenium.WebDriver, wffn waitForFn, d time.Duration) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	var found bool
	for !found {
		select {
		case <-ctx.Done():
			goto END
		default:
		}
		found = wffn(t, wd)
	}

END:
	if !found {
		screenshot(t, wd)
	}
	require.True(t, found)
}

func eqText(by, value, txt string) waitForFn {
	return func(t *testing.T, wd selenium.WebDriver) bool {
		we, err := wd.FindElement(by, value)
		if err != nil {
			return false
		}
		weTxt, err := we.Text()
		if err != nil {
			return false
		}
		return weTxt == txt
	}
}

func containsText(by, value, substr string) waitForFn {
	return func(t *testing.T, wd selenium.WebDriver) bool {
		we, err := wd.FindElement(by, value)
		if err != nil {
			return false
		}
		weTxt, err := we.Text()
		if err != nil {
			return false
		}
		return len(weTxt) > 0 && contains(weTxt, substr)
	}
}

func elementExists(by, value string) waitForFn {
	return func(t *testing.T, wd selenium.WebDriver) bool {
		_, err := wd.FindElement(by, value)
		return err == nil
	}
}

func elementsCount(by, value string, count int) waitForFn {
	return func(t *testing.T, wd selenium.WebDriver) bool {
		elems, err := wd.FindElements(by, value)
		if err != nil {
			return false
		}
		return len(elems) == count
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstr(s, substr)
}

func searchSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func screenshot(t *testing.T, wd selenium.WebDriver) {
	t.Helper()
	b, err := wd.Screenshot()
	if err != nil {
		return
	}
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return
	}
	f, err := os.Create("screenshot.png")
	if err != nil {
		return
	}
	defer f.Close()
	png.Encode(f, img)
}

// loginViaLocalStorage injects a JWT token and user data into localStorage
// to simulate an authenticated session without going through GitHub OAuth.
func loginViaLocalStorage(t *testing.T, wd selenium.WebDriver, serverURL, token, username string) {
	t.Helper()

	// First navigate to the server so we're on the right origin
	err := wd.Get(serverURL)
	require.NoError(t, err)
	waitFor(t, wd, elementExists(selenium.ByCSSSelector, "#app"), 10*time.Second)

	user := map[string]string{
		"Username":  username,
		"AvatarURL": "https://example.com/" + username + ".png",
	}
	userJSON, _ := json.Marshal(user)

	_, err = wd.ExecuteScript(fmt.Sprintf(`
		localStorage.setItem('pkreg-token', '%s');
		localStorage.setItem('pkreg-user', '%s');
	`, token, string(userJSON)), nil)
	require.NoError(t, err)
}
