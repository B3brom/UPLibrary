package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"clothingStore/internal/db"
	"clothingStore/internal/models"
	"clothingStore/internal/repository"
)

func TestRegisterAndLogin(t *testing.T) {
	database, err := db.Connect(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	database.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}

	handler := NewAuthHandler(repository.NewUserRepo(database), "test-secret", time.Hour)
	register := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(
		`{"name":"Test User","email":"TEST@example.com","password":"password123","role_id":"00000000-0000-0000-0000-000000000004"}`,
	))
	registerResponse := httptest.NewRecorder()
	handler.Register(registerResponse, register)
	if registerResponse.Code != http.StatusOK {
		t.Fatalf("register status = %d, body = %s", registerResponse.Code, registerResponse.Body.String())
	}

	var session loginResp
	if err := json.Unmarshal(registerResponse.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if session.Token == "" || session.User.ID == "" {
		t.Fatalf("registration returned incomplete session: %+v", session)
	}
	if session.User.RoleID != models.RoleClient {
		t.Fatalf("registered role = %q, want client role %q", session.User.RoleID, models.RoleClient)
	}

	duplicate := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(
		`{"name":"Other User","email":"test@example.com","password":"password123"}`,
	))
	duplicateResponse := httptest.NewRecorder()
	handler.Register(duplicateResponse, duplicate)
	if duplicateResponse.Code != http.StatusConflict {
		t.Fatalf("duplicate registration status = %d, want %d", duplicateResponse.Code, http.StatusConflict)
	}

	login := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(
		`{"email":"test@example.com","password":"password123"}`,
	))
	loginResponse := httptest.NewRecorder()
	handler.Login(loginResponse, login)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", loginResponse.Code, loginResponse.Body.String())
	}

	badLogin := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(
		`{"email":"test@example.com","password":"wrong-password"}`,
	))
	badLoginResponse := httptest.NewRecorder()
	handler.Login(badLoginResponse, badLogin)
	if badLoginResponse.Code != http.StatusUnauthorized {
		t.Fatalf("invalid password status = %d, want %d", badLoginResponse.Code, http.StatusUnauthorized)
	}
}
