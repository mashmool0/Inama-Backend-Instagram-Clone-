package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type e2eTokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type e2eProfile struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatar_url"`
}

type e2eNotification struct {
	ID      string `json:"id"`
	ActorID string `json:"actor_id"`
	IsRead  bool   `json:"is_read"`
}

type e2eNotificationPage struct {
	Notifications []e2eNotification `json:"notifications"`
}

func TestFourServiceE2E(t *testing.T) {
	baseURL := strings.TrimRight(os.Getenv("E2E_GATEWAY_URL"), "/")
	if baseURL == "" {
		t.Skip("E2E_GATEWAY_URL is not set")
	}

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	usernameA, usernameB := "e2e_a_"+suffix, "e2e_b_"+suffix
	pairA := e2eRegister(t, baseURL, usernameA+"@example.com", usernameA)
	pairB := e2eRegister(t, baseURL, usernameB+"@example.com", usernameB)
	userIDA, userIDB := e2eTokenSubject(t, pairA.AccessToken), e2eTokenSubject(t, pairB.AccessToken)

	profileA := e2eWaitForProfile(t, baseURL, pairA.AccessToken, "/users/"+userIDA, usernameA)
	_ = e2eWaitForProfile(t, baseURL, pairB.AccessToken, "/users/username/"+usernameB, usernameB)
	if profileA.ID != userIDA {
		t.Fatalf("profile id = %q, auth subject = %q", profileA.ID, userIDA)
	}

	var login e2eTokenPair
	e2eJSON(t, http.MethodPost, baseURL+"/auth/login", "", map[string]string{
		"identifier": usernameA,
		"password":   "password123",
	}, http.StatusOK, &login)
	var refreshed e2eTokenPair
	e2eJSON(t, http.MethodPost, baseURL+"/auth/refresh", "", map[string]string{
		"refresh_token": login.RefreshToken,
	}, http.StatusOK, &refreshed)
	if refreshed.RefreshToken == login.RefreshToken {
		t.Fatal("refresh token was not rotated")
	}

	var updated e2eProfile
	e2eJSON(t, http.MethodPatch, baseURL+"/users/me", pairA.AccessToken, map[string]string{
		"bio":        "E2E profile",
		"avatar_url": "https://example.com/avatar.png",
	}, http.StatusOK, &updated)
	if updated.Bio != "E2E profile" || updated.AvatarURL != "https://example.com/avatar.png" {
		t.Fatalf("updated profile = %+v", updated)
	}

	newUsernameA := "renamed_" + suffix
	e2eJSON(t, http.MethodPatch, baseURL+"/auth/me/username", pairA.AccessToken, map[string]string{
		"username": newUsernameA,
	}, http.StatusOK, nil)
	_ = e2eWaitForProfile(t, baseURL, pairA.AccessToken, "/users/username/"+newUsernameA, newUsernameA)

	e2eJSON(t, http.MethodPost, baseURL+"/users/"+userIDB+"/follow", pairA.AccessToken, nil, http.StatusOK, nil)
	// Idempotent replay at the API level must not publish a second follow event.
	e2eJSON(t, http.MethodPost, baseURL+"/users/"+userIDB+"/follow", pairA.AccessToken, nil, http.StatusOK, nil)
	notification := e2eWaitForFollowNotification(t, baseURL, pairB.AccessToken, userIDA)

	e2eJSON(t, http.MethodPost, baseURL+"/notifications/"+notification.ID+"/read", pairA.AccessToken, nil, http.StatusNotFound, nil)
	e2eJSON(t, http.MethodPost, baseURL+"/notifications/"+notification.ID+"/read", pairB.AccessToken, nil, http.StatusOK, nil)
	e2eJSON(t, http.MethodPost, baseURL+"/notifications/read-all", pairB.AccessToken, nil, http.StatusOK, nil)
	e2eJSON(t, http.MethodGet, baseURL+"/notifications", "", nil, http.StatusUnauthorized, nil)
}

func e2eRegister(t *testing.T, baseURL, email, username string) e2eTokenPair {
	t.Helper()
	var pair e2eTokenPair
	e2eJSON(t, http.MethodPost, baseURL+"/auth/register", "", map[string]string{
		"email":    email,
		"username": username,
		"password": "password123",
	}, http.StatusCreated, &pair)
	return pair
}

func e2eWaitForProfile(t *testing.T, baseURL, token, path, username string) e2eProfile {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		var profile e2eProfile
		statusCode := e2eRequest(t, http.MethodGet, baseURL+path, token, nil, &profile)
		if statusCode == http.StatusOK && profile.Username == username {
			return profile
		}
		if statusCode != http.StatusNotFound && statusCode != http.StatusOK {
			t.Fatalf("wait for profile status = %d", statusCode)
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("profile %q did not converge", username)
	return e2eProfile{}
}

func e2eWaitForFollowNotification(t *testing.T, baseURL, token, actorID string) e2eNotification {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		var page e2eNotificationPage
		statusCode := e2eRequest(t, http.MethodGet, baseURL+"/notifications", token, nil, &page)
		if statusCode != http.StatusOK {
			t.Fatalf("get notifications status = %d", statusCode)
		}
		matches := make([]e2eNotification, 0, 1)
		for _, notification := range page.Notifications {
			if notification.ActorID == actorID {
				matches = append(matches, notification)
			}
		}
		if len(matches) == 1 {
			return matches[0]
		}
		if len(matches) > 1 {
			t.Fatalf("duplicate follow notifications = %d", len(matches))
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("follow notification did not arrive")
	return e2eNotification{}
}

func e2eTokenSubject(t *testing.T, rawToken string) string {
	t.Helper()
	claims := &jwt.RegisteredClaims{}
	if _, _, err := new(jwt.Parser).ParseUnverified(rawToken, claims); err != nil {
		t.Fatalf("ParseUnverified() error = %v", err)
	}
	if claims.Subject == "" {
		t.Fatal("access token has no subject")
	}
	return claims.Subject
}

func e2eJSON(t *testing.T, method, url, token string, body any, expectedStatus int, target any) {
	t.Helper()
	statusCode := e2eRequest(t, method, url, token, body, target)
	if statusCode != expectedStatus {
		t.Fatalf("%s %s status = %d, want %d", method, url, statusCode, expectedStatus)
	}
}

func e2eRequest(t *testing.T, method, url, token string, body, target any) int {
	t.Helper()
	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		payload = bytes.NewReader(encoded)
	}
	request, err := http.NewRequest(method, url, payload)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if target != nil && response.StatusCode >= 200 && response.StatusCode < 300 {
		if err := json.Unmarshal(responseBody, target); err != nil {
			t.Fatalf("Unmarshal(%s) error = %v; body=%s", url, err, responseBody)
		}
	}
	return response.StatusCode
}
