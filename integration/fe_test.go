//go:build integration

package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tebeka/selenium"
)

func TestFrontend(t *testing.T) {
	env := setupTestEnv(t)
	wd := getRemote(t)
	wd.ResizeWindow("", 1500, 1500)

	serverURL := env.server.URL

	// Seed users and publish some plugins for testing
	_, aliceToken := seedUser(t, env, "alice")
	seedUser(t, env, "bob")

	// Publish plugins via API so we have data to browse
	manifest := sampleManifest("postgres", "1.0.0", "resource_type")
	readme := []byte("# PostgreSQL Resource Type\nManages PostgreSQL databases.")
	resp := publishManifest(t, env, aliceToken, "alice", "postgres", "1.0.0", manifest, readme)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	manifest2 := sampleManifest("redis", "0.1.0", "service_type")
	resp = publishManifest(t, env, aliceToken, "alice", "redis", "0.1.0", manifest2, nil)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Add tags to postgres
	doAuthPatch(t, serverURL, "/api/plugins/alice/postgres/tags", aliceToken, map[string]interface{}{
		"tags": []string{"database", "sql"},
	}).Body.Close()

	t.Run("HomePage", func(t *testing.T) {
		t.Run("RendersHeroSection", func(t *testing.T) {
			err := wd.Get(serverURL)
			require.NoError(t, err)

			waitFor(t, wd, containsText(selenium.ByCSSSelector, ".search-hero h1", "Plugin Registry"), 10*time.Second)

			subtitle, err := wd.FindElement(selenium.ByCSSSelector, ".search-hero p")
			require.NoError(t, err)
			txt, err := subtitle.Text()
			require.NoError(t, err)
			assert.Equal(t, "Discover and share plugins for your infrastructure", txt)
		})

		t.Run("RendersSearchResults", func(t *testing.T) {
			err := wd.Get(serverURL)
			require.NoError(t, err)

			// Wait for plugin cards to appear
			waitFor(t, wd, elementsCount(selenium.ByCSSSelector, ".pkreg-card", 2), 10*time.Second)

			cards, err := wd.FindElements(selenium.ByCSSSelector, ".pkreg-card")
			require.NoError(t, err)
			assert.Equal(t, 2, len(cards))
		})

		t.Run("RendersKindTabs", func(t *testing.T) {
			err := wd.Get(serverURL)
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".kind-tabs"), 10*time.Second)

			tabs, err := wd.FindElements(selenium.ByCSSSelector, ".kind-tab")
			require.NoError(t, err)
			assert.Equal(t, 6, len(tabs)) // All, Resource, Runner, Service, Secret, Notification
		})

		t.Run("KindFilterFilters", func(t *testing.T) {
			err := wd.Get(serverURL)
			require.NoError(t, err)

			waitFor(t, wd, elementsCount(selenium.ByCSSSelector, ".pkreg-card", 2), 10*time.Second)

			// Click on "Resource" tab
			resourceTab, err := wd.FindElement(selenium.ByCSSSelector, `.kind-tab[data-kind="resource_type"]`)
			require.NoError(t, err)
			err = resourceTab.Click()
			require.NoError(t, err)

			// Should only show 1 card (postgres is resource_type)
			waitFor(t, wd, elementsCount(selenium.ByCSSSelector, ".pkreg-card", 1), 10*time.Second)

			// Click on "Service" tab
			serviceTab, err := wd.FindElement(selenium.ByCSSSelector, `.kind-tab[data-kind="service_type"]`)
			require.NoError(t, err)
			err = serviceTab.Click()
			require.NoError(t, err)

			// Should only show 1 card (redis is service_type)
			waitFor(t, wd, elementsCount(selenium.ByCSSSelector, ".pkreg-card", 1), 10*time.Second)

			// Click "All" tab to reset
			allTab, err := wd.FindElement(selenium.ByCSSSelector, `.kind-tab[data-kind=""]`)
			require.NoError(t, err)
			err = allTab.Click()
			require.NoError(t, err)

			waitFor(t, wd, elementsCount(selenium.ByCSSSelector, ".pkreg-card", 2), 10*time.Second)
		})

		t.Run("HeroSearchSubmit", func(t *testing.T) {
			err := wd.Get(serverURL)
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".hero-search-form"), 10*time.Second)

			input, err := wd.FindElement(selenium.ByCSSSelector, ".hero-search-form input")
			require.NoError(t, err)
			input.SendKeys("postgres")
			input.SendKeys(selenium.EnterKey)

			// Should show only the postgres card
			waitFor(t, wd, elementsCount(selenium.ByCSSSelector, ".pkreg-card", 1), 10*time.Second)

			card, err := wd.FindElement(selenium.ByCSSSelector, ".card-title a")
			require.NoError(t, err)
			txt, err := card.Text()
			require.NoError(t, err)
			assert.Equal(t, "postgres", txt)
		})
	})

	t.Run("Navbar", func(t *testing.T) {
		t.Run("SearchSubmit", func(t *testing.T) {
			err := wd.Get(serverURL)
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".navbar-search-input"), 10*time.Second)

			input, err := wd.FindElement(selenium.ByCSSSelector, ".navbar-search-input")
			require.NoError(t, err)
			input.SendKeys("redis")
			input.SendKeys(selenium.EnterKey)

			waitFor(t, wd, elementsCount(selenium.ByCSSSelector, ".pkreg-card", 1), 10*time.Second)

			card, err := wd.FindElement(selenium.ByCSSSelector, ".card-title a")
			require.NoError(t, err)
			txt, err := card.Text()
			require.NoError(t, err)
			assert.Equal(t, "redis", txt)
		})

		t.Run("LogoNavigatesHome", func(t *testing.T) {
			err := wd.Get(serverURL + "/tags")
			require.NoError(t, err)

			waitFor(t, wd, containsText(selenium.ByCSSSelector, "h1", "Tags"), 10*time.Second)

			logo, err := wd.FindElement(selenium.ByCSSSelector, ".navbar-brand")
			require.NoError(t, err)
			err = logo.Click()
			require.NoError(t, err)

			waitFor(t, wd, containsText(selenium.ByCSSSelector, ".search-hero h1", "Plugin Registry"), 10*time.Second)
		})

		t.Run("ThemeToggle", func(t *testing.T) {
			err := wd.Get(serverURL)
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".theme-toggle"), 10*time.Second)

			// Get initial theme
			initialTheme, err := wd.ExecuteScript(`return document.documentElement.getAttribute('data-theme')`, nil)
			require.NoError(t, err)

			// Click theme toggle
			toggle, err := wd.FindElement(selenium.ByCSSSelector, ".theme-toggle")
			require.NoError(t, err)
			err = toggle.Click()
			require.NoError(t, err)

			// Wait a moment for the toggle to take effect
			time.Sleep(500 * time.Millisecond)

			newTheme, err := wd.ExecuteScript(`return document.documentElement.getAttribute('data-theme')`, nil)
			require.NoError(t, err)

			assert.NotEqual(t, initialTheme, newTheme)
		})

		t.Run("TagsLink", func(t *testing.T) {
			err := wd.Get(serverURL)
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, `a[href="/tags"]`), 10*time.Second)

			link, err := wd.FindElement(selenium.ByCSSSelector, `a.nav-link[href="/tags"]`)
			require.NoError(t, err)
			err = link.Click()
			require.NoError(t, err)

			waitFor(t, wd, containsText(selenium.ByCSSSelector, "h1", "Tags"), 10*time.Second)
		})

		t.Run("SignInLinkWhenLoggedOut", func(t *testing.T) {
			// Clear auth
			wd.ExecuteScript(`localStorage.removeItem('pkreg-token'); localStorage.removeItem('pkreg-user');`, nil)

			err := wd.Get(serverURL)
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, `a[href="/login"]`), 10*time.Second)

			link, err := wd.FindElement(selenium.ByCSSSelector, `a.nav-link[href="/login"]`)
			require.NoError(t, err)
			txt, err := link.Text()
			require.NoError(t, err)
			assert.Contains(t, txt, "Sign In")
		})

		t.Run("AuthenticatedNavLinks", func(t *testing.T) {
			loginViaLocalStorage(t, wd, serverURL, aliceToken, "alice")

			err := wd.Get(serverURL)
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, `a.nav-link[href="/me/types"]`), 10*time.Second)

			myPlugins, err := wd.FindElement(selenium.ByCSSSelector, `a.nav-link[href="/me/types"]`)
			require.NoError(t, err)
			txt, err := myPlugins.Text()
			require.NoError(t, err)
			assert.Contains(t, txt, "My Plugins")

			tokens, err := wd.FindElement(selenium.ByCSSSelector, `a.nav-link[href="/me/tokens"]`)
			require.NoError(t, err)
			txt, err = tokens.Text()
			require.NoError(t, err)
			assert.Contains(t, txt, "Tokens")
		})

		t.Run("LogoutClearsSession", func(t *testing.T) {
			loginViaLocalStorage(t, wd, serverURL, aliceToken, "alice")

			err := wd.Get(serverURL)
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".dropdown-toggle"), 10*time.Second)

			// Open dropdown
			dropdown, err := wd.FindElement(selenium.ByCSSSelector, ".dropdown-toggle")
			require.NoError(t, err)
			err = dropdown.Click()
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".logout-btn"), 5*time.Second)

			logoutBtn, err := wd.FindElement(selenium.ByCSSSelector, ".logout-btn")
			require.NoError(t, err)
			err = logoutBtn.Click()
			require.NoError(t, err)

			// Should now see "Sign In" link
			waitFor(t, wd, elementExists(selenium.ByCSSSelector, `a[href="/login"]`), 10*time.Second)

			// Verify localStorage is cleared
			tokenVal, err := wd.ExecuteScript(`return localStorage.getItem('pkreg-token')`, nil)
			require.NoError(t, err)
			assert.Nil(t, tokenVal)
		})
	})

	t.Run("PluginDetailPage", func(t *testing.T) {
		t.Run("RendersPluginInfo", func(t *testing.T) {
			err := wd.Get(serverURL + "/plugins/alice/postgres")
			require.NoError(t, err)

			waitFor(t, wd, eqText(selenium.ByCSSSelector, ".type-header h1", "postgres"), 10*time.Second)

			// Check namespace link
			nsLink, err := wd.FindElement(selenium.ByCSSSelector, ".namespace-link a")
			require.NoError(t, err)
			nsTxt, err := nsLink.Text()
			require.NoError(t, err)
			assert.Contains(t, nsTxt, "alice")

			// Check kind badge
			badge, err := wd.FindElement(selenium.ByCSSSelector, ".kind-badge")
			require.NoError(t, err)
			badgeTxt, err := badge.Text()
			require.NoError(t, err)
			assert.Contains(t, badgeTxt, "Resource Type")

			// Check description
			desc, err := wd.FindElement(selenium.ByCSSSelector, ".type-header p")
			require.NoError(t, err)
			descTxt, err := desc.Text()
			require.NoError(t, err)
			assert.Contains(t, descTxt, "test resource_type")
		})

		t.Run("RendersTags", func(t *testing.T) {
			err := wd.Get(serverURL + "/plugins/alice/postgres")
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".tag-cloud"), 10*time.Second)

			tags, err := wd.FindElements(selenium.ByCSSSelector, ".tag-pill")
			require.NoError(t, err)
			assert.Equal(t, 2, len(tags)) // "database" and "sql"
		})

		t.Run("VersionsTab", func(t *testing.T) {
			err := wd.Get(serverURL + "/plugins/alice/postgres")
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, `.detail-tab[data-tab="versions"]`), 10*time.Second)

			// Versions tab should be active by default
			versionsTab, err := wd.FindElement(selenium.ByCSSSelector, `.detail-tab[data-tab="versions"]`)
			require.NoError(t, err)
			cls, err := versionsTab.GetAttribute("class")
			require.NoError(t, err)
			assert.Contains(t, cls, "active")

			// Check version row
			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".version-row"), 10*time.Second)

			verNum, err := wd.FindElement(selenium.ByCSSSelector, ".version-number")
			require.NoError(t, err)
			verTxt, err := verNum.Text()
			require.NoError(t, err)
			assert.Equal(t, "1.0.0", verTxt)
		})

		t.Run("ReadmeTab", func(t *testing.T) {
			err := wd.Get(serverURL + "/plugins/alice/postgres")
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, `.detail-tab[data-tab="readme"]`), 10*time.Second)

			readmeTab, err := wd.FindElement(selenium.ByCSSSelector, `.detail-tab[data-tab="readme"]`)
			require.NoError(t, err)
			err = readmeTab.Click()
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".readme-content"), 5*time.Second)

			readme, err := wd.FindElement(selenium.ByCSSSelector, ".readme-content pre")
			require.NoError(t, err)
			readmeTxt, err := readme.Text()
			require.NoError(t, err)
			assert.Contains(t, readmeTxt, "PostgreSQL Resource Type")
		})

		t.Run("ContentTab", func(t *testing.T) {
			err := wd.Get(serverURL + "/plugins/alice/postgres")
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, `.detail-tab[data-tab="content"]`), 10*time.Second)

			contentTab, err := wd.FindElement(selenium.ByCSSSelector, `.detail-tab[data-tab="content"]`)
			require.NoError(t, err)
			err = contentTab.Click()
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".hcl-content"), 5*time.Second)

			content, err := wd.FindElement(selenium.ByCSSSelector, ".hcl-content")
			require.NoError(t, err)
			contentTxt, err := content.Text()
			require.NoError(t, err)
			assert.Contains(t, contentTxt, "resource_type")
		})

		t.Run("TagPillNavigatesToTagDetail", func(t *testing.T) {
			err := wd.Get(serverURL + "/plugins/alice/postgres")
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".tag-pill"), 10*time.Second)

			tag, err := wd.FindElement(selenium.ByCSSSelector, ".tag-pill")
			require.NoError(t, err)
			err = tag.Click()
			require.NoError(t, err)

			// Should navigate to tag detail page
			waitFor(t, wd, containsText(selenium.ByCSSSelector, "h1", "database"), 10*time.Second)
		})

		t.Run("NamespaceLinkNavigates", func(t *testing.T) {
			err := wd.Get(serverURL + "/plugins/alice/postgres")
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".namespace-link a"), 10*time.Second)

			nsLink, err := wd.FindElement(selenium.ByCSSSelector, ".namespace-link a")
			require.NoError(t, err)
			err = nsLink.Click()
			require.NoError(t, err)

			// Should navigate to namespace page
			waitFor(t, wd, containsText(selenium.ByCSSSelector, ".namespace-header h1", "alice"), 10*time.Second)
		})
	})

	t.Run("NamespacePage", func(t *testing.T) {
		t.Run("RendersPlugins", func(t *testing.T) {
			err := wd.Get(serverURL + "/plugins/alice")
			require.NoError(t, err)

			waitFor(t, wd, containsText(selenium.ByCSSSelector, ".namespace-header h1", "alice"), 10*time.Second)

			// Check plugin count text
			countText, err := wd.FindElement(selenium.ByCSSSelector, ".namespace-header .text-muted")
			require.NoError(t, err)
			txt, err := countText.Text()
			require.NoError(t, err)
			assert.Contains(t, txt, "2 plugins")

			// Check cards
			waitFor(t, wd, elementsCount(selenium.ByCSSSelector, ".pkreg-card", 2), 10*time.Second)
		})

		t.Run("CardNavigatesToDetail", func(t *testing.T) {
			err := wd.Get(serverURL + "/plugins/alice")
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".card-title a"), 10*time.Second)

			link, err := wd.FindElement(selenium.ByCSSSelector, ".card-title a")
			require.NoError(t, err)
			err = link.Click()
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".type-header h1"), 10*time.Second)
		})

		t.Run("EmptyNamespace", func(t *testing.T) {
			err := wd.Get(serverURL + "/plugins/bob")
			require.NoError(t, err)

			waitFor(t, wd, containsText(selenium.ByCSSSelector, ".namespace-header h1", "bob"), 10*time.Second)
			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".empty-state"), 10*time.Second)
		})
	})

	t.Run("TagsPage", func(t *testing.T) {
		t.Run("RendersTagCloud", func(t *testing.T) {
			err := wd.Get(serverURL + "/tags")
			require.NoError(t, err)

			waitFor(t, wd, containsText(selenium.ByCSSSelector, "h1", "Tags"), 10*time.Second)

			tags, err := wd.FindElements(selenium.ByCSSSelector, ".tag-pill")
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(tags), 2) // at least "database" and "sql"
		})

		t.Run("TagCountBadge", func(t *testing.T) {
			err := wd.Get(serverURL + "/tags")
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".tag-count"), 10*time.Second)

			counts, err := wd.FindElements(selenium.ByCSSSelector, ".tag-count")
			require.NoError(t, err)
			assert.Greater(t, len(counts), 0)
		})

		t.Run("TagPillNavigatesToDetail", func(t *testing.T) {
			err := wd.Get(serverURL + "/tags")
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".tag-pill"), 10*time.Second)

			tag, err := wd.FindElement(selenium.ByCSSSelector, ".tag-pill")
			require.NoError(t, err)
			err = tag.Click()
			require.NoError(t, err)

			// Should navigate to tag detail page
			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".type-cards-grid"), 10*time.Second)
		})
	})

	t.Run("TagDetailPage", func(t *testing.T) {
		t.Run("RendersPluginsWithTag", func(t *testing.T) {
			err := wd.Get(serverURL + "/tags/database")
			require.NoError(t, err)

			waitFor(t, wd, containsText(selenium.ByCSSSelector, "h1", "database"), 10*time.Second)

			// Should show plugin count
			countText, err := wd.FindElement(selenium.ByCSSSelector, "p.text-muted")
			require.NoError(t, err)
			txt, err := countText.Text()
			require.NoError(t, err)
			assert.Contains(t, txt, "1 plugin")

			// Should show the postgres card
			waitFor(t, wd, elementsCount(selenium.ByCSSSelector, ".pkreg-card", 1), 10*time.Second)
		})

		t.Run("BackLinkToAllTags", func(t *testing.T) {
			err := wd.Get(serverURL + "/tags/database")
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, `a[href="/tags"]`), 10*time.Second)

			backLink, err := wd.FindElement(selenium.ByCSSSelector, `a.text-muted[href="/tags"]`)
			require.NoError(t, err)
			err = backLink.Click()
			require.NoError(t, err)

			waitFor(t, wd, containsText(selenium.ByCSSSelector, "h1", "Tags"), 10*time.Second)
		})

		t.Run("EmptyTag", func(t *testing.T) {
			err := wd.Get(serverURL + "/tags/nonexistent")
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".empty-state"), 10*time.Second)
		})
	})

	t.Run("LoginPage", func(t *testing.T) {
		t.Run("RendersLoginForm", func(t *testing.T) {
			// Clear auth first
			wd.ExecuteScript(`localStorage.removeItem('pkreg-token'); localStorage.removeItem('pkreg-user');`, nil)

			err := wd.Get(serverURL + "/login")
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".auth-container"), 10*time.Second)

			heading, err := wd.FindElement(selenium.ByCSSSelector, ".auth-container h2")
			require.NoError(t, err)
			txt, err := heading.Text()
			require.NoError(t, err)
			assert.Contains(t, txt, "Sign in to pkreg")

			// GitHub button should be present
			ghBtn, err := wd.FindElement(selenium.ByCSSSelector, ".btn-github")
			require.NoError(t, err)
			btnTxt, err := ghBtn.Text()
			require.NoError(t, err)
			assert.Contains(t, btnTxt, "Sign in with GitHub")
		})
	})

	t.Run("MyPluginsPage", func(t *testing.T) {
		t.Run("RedirectsToLoginWhenNotAuthenticated", func(t *testing.T) {
			// Clear auth
			wd.ExecuteScript(`localStorage.removeItem('pkreg-token'); localStorage.removeItem('pkreg-user');`, nil)

			err := wd.Get(serverURL + "/me/types")
			require.NoError(t, err)

			// Should redirect to login
			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".auth-container"), 10*time.Second)
		})

		t.Run("RendersMyPlugins", func(t *testing.T) {
			loginViaLocalStorage(t, wd, serverURL, aliceToken, "alice")

			err := wd.Get(serverURL + "/me/types")
			require.NoError(t, err)

			waitFor(t, wd, containsText(selenium.ByCSSSelector, "h1", "My Plugins"), 10*time.Second)

			// Should show the username
			waitFor(t, wd, containsText(selenium.ByCSSSelector, ".fw-semibold", "alice"), 10*time.Second)

			// Should show the plugins table
			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".pkreg-table"), 10*time.Second)

			rows, err := wd.FindElements(selenium.ByCSSSelector, ".pkreg-table tbody tr")
			require.NoError(t, err)
			assert.Equal(t, 2, len(rows)) // postgres and redis
		})

		t.Run("PluginLinkNavigatesToDetail", func(t *testing.T) {
			loginViaLocalStorage(t, wd, serverURL, aliceToken, "alice")

			err := wd.Get(serverURL + "/me/types")
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".pkreg-table tbody a"), 10*time.Second)

			link, err := wd.FindElement(selenium.ByCSSSelector, ".pkreg-table tbody a")
			require.NoError(t, err)
			err = link.Click()
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".type-header h1"), 10*time.Second)
		})
	})

	t.Run("TokensPage", func(t *testing.T) {
		t.Run("RedirectsToLoginWhenNotAuthenticated", func(t *testing.T) {
			// Clear auth
			wd.ExecuteScript(`localStorage.removeItem('pkreg-token'); localStorage.removeItem('pkreg-user');`, nil)

			err := wd.Get(serverURL + "/me/tokens")
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".auth-container"), 10*time.Second)
		})

		t.Run("RendersTokenPage", func(t *testing.T) {
			loginViaLocalStorage(t, wd, serverURL, aliceToken, "alice")

			err := wd.Get(serverURL + "/me/tokens")
			require.NoError(t, err)

			waitFor(t, wd, containsText(selenium.ByCSSSelector, "h1", "API Tokens"), 10*time.Second)

			// Create token form should be present
			form, err := wd.FindElement(selenium.ByCSSSelector, ".create-token-form")
			require.NoError(t, err)
			assert.NotNil(t, form)

			// Token name input
			nameInput, err := wd.FindElement(selenium.ByCSSSelector, ".token-name-input")
			require.NoError(t, err)
			assert.NotNil(t, nameInput)

			// Initially should show empty state (no tokens)
			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".empty-state"), 10*time.Second)
		})

		t.Run("CreateToken", func(t *testing.T) {
			loginViaLocalStorage(t, wd, serverURL, aliceToken, "alice")

			err := wd.Get(serverURL + "/me/tokens")
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".create-token-form"), 10*time.Second)

			// Fill in token name
			nameInput, err := wd.FindElement(selenium.ByCSSSelector, ".token-name-input")
			require.NoError(t, err)
			nameInput.SendKeys("ci-deploy")

			// Fill in namespace (optional)
			nsInput, err := wd.FindElement(selenium.ByCSSSelector, ".token-namespace-input")
			require.NoError(t, err)
			nsInput.SendKeys("alice")

			// Submit form
			submitBtn, err := wd.FindElement(selenium.ByCSSSelector, ".create-token-form button[type=submit]")
			require.NoError(t, err)
			err = submitBtn.Click()
			require.NoError(t, err)

			// Should show the token display
			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".token-display"), 15*time.Second)

			tokenDisplay, err := wd.FindElement(selenium.ByCSSSelector, ".token-display")
			require.NoError(t, err)
			tokenTxt, err := tokenDisplay.Text()
			require.NoError(t, err)
			assert.NotEmpty(t, tokenTxt)

			// Should show token in the table
			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".pkreg-table"), 10*time.Second)

			rows, err := wd.FindElements(selenium.ByCSSSelector, ".pkreg-table tbody tr")
			require.NoError(t, err)
			assert.Equal(t, 1, len(rows))
		})

		t.Run("RevokeToken", func(t *testing.T) {
			loginViaLocalStorage(t, wd, serverURL, aliceToken, "alice")

			// First create a token via API so we have one to revoke
			resp := doAuthPost(t, serverURL, "/api/me/tokens", aliceToken, map[string]interface{}{
				"name":      "to-revoke",
				"namespace": "alice",
			})
			resp.Body.Close()

			err := wd.Get(serverURL + "/me/tokens")
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".revoke-token-btn"), 15*time.Second)

			// Count tokens before revoke
			rowsBefore, err := wd.FindElements(selenium.ByCSSSelector, ".pkreg-table tbody tr")
			require.NoError(t, err)
			countBefore := len(rowsBefore)

			// Click revoke - need to handle confirm dialog
			revokeBtn, err := wd.FindElement(selenium.ByCSSSelector, ".revoke-token-btn")
			require.NoError(t, err)

			// Override confirm to auto-accept
			wd.ExecuteScript(`window.confirm = function() { return true; }`, nil)

			err = revokeBtn.Click()
			require.NoError(t, err)

			// Wait for the table to update (one less row)
			if countBefore > 1 {
				waitFor(t, wd, elementsCount(selenium.ByCSSSelector, ".pkreg-table tbody tr", countBefore-1), 10*time.Second)
			} else {
				// If it was the last token, empty state should appear
				waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".empty-state"), 10*time.Second)
			}
		})
	})

	t.Run("SearchResultNavigation", func(t *testing.T) {
		t.Run("ClickCardNavigatesToDetail", func(t *testing.T) {
			err := wd.Get(serverURL)
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".card-title a"), 10*time.Second)

			link, err := wd.FindElement(selenium.ByCSSSelector, ".card-title a")
			require.NoError(t, err)
			err = link.Click()
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".type-header h1"), 10*time.Second)
		})

		t.Run("ClickNamespaceNavigates", func(t *testing.T) {
			err := wd.Get(serverURL)
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".pkreg-card .text-muted a"), 10*time.Second)

			nsLink, err := wd.FindElement(selenium.ByCSSSelector, ".pkreg-card .text-muted a")
			require.NoError(t, err)
			err = nsLink.Click()
			require.NoError(t, err)

			waitFor(t, wd, elementExists(selenium.ByCSSSelector, ".namespace-header"), 10*time.Second)
		})
	})
}
