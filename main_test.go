package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCaptchaRequestIncludesWidgetRecordID(t *testing.T) {
	client := &infraiClient{key: "test", http: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["widget_record_id"] != "widget-123" || body["token"] != "token-123" {
			t.Fatalf("unexpected captcha request: %v", body)
		}
		return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
	})}}
	if err := client.captcha("token-123", "widget-123"); err == nil || !strings.Contains(err.Error(), "captcha rejected") {
		t.Fatalf("unexpected captcha response: %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSignupThenReport(t *testing.T) {
	for _, tc := range []struct{ name, email string }{{"first learner", "l@example.com"}} {
		t.Run(tc.name, func(t *testing.T) {
			a := &app{learners: map[string]learner{}, sessions: map[string]string{}}
			req := httptest.NewRequest("POST", "/signup", strings.NewReader(`{"email":"`+tc.email+`","password":"p","name":"Lin"}`))
			rec := httptest.NewRecorder()
			a.signup(rec, req)
			if rec.Code != 200 {
				t.Fatalf("signup status %d", rec.Code)
			}
			rr := httptest.NewRecorder()
			a.report(rr, httptest.NewRequest("GET", "/report", nil))
			if !strings.Contains(rr.Body.String(), `"learners":1`) {
				t.Fatal(rr.Body.String())
			}
		})
	}
}
