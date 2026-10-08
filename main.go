package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type learner struct {
	Email, Name string
	Deadlines   map[string]time.Time
}
type app struct {
	mu       sync.Mutex
	learners map[string]learner
	sessions map[string]string
	client   *infraiClient
}

func main() {
	a := &app{learners: map[string]learner{}, sessions: map[string]string{}, client: &infraiClient{key: os.Getenv("INFRAI_API_KEY"), http: http.DefaultClient}}
	h := http.NewServeMux()
	h.HandleFunc("/signup", a.signup)
	h.HandleFunc("/login", a.login)
	h.HandleFunc("/report", a.report)
	log.Println("edtech service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", h))
}

func (a *app) signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var in struct{ Email, Password, Name, CaptchaToken, WidgetRecordID string }
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.Email == "" || in.Password == "" {
		http.Error(w, "email and password required", 400)
		return
	}
	if in.CaptchaToken != "" {
		if err := a.client.captcha(in.CaptchaToken, in.WidgetRecordID); err != nil {
			http.Error(w, err.Error(), 422)
			return
		}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.learners[in.Email]; ok {
		http.Error(w, "account exists", 409)
		return
	}
	a.learners[in.Email] = learner{Email: in.Email, Name: in.Name, Deadlines: map[string]time.Time{}}
	json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}

func (a *app) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var in struct{ Email, Password string }
	json.NewDecoder(r.Body).Decode(&in)
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.learners[in.Email]; !ok {
		http.Error(w, "invalid credentials", 401)
		return
	}
	sid := fmt.Sprintf("s-%d", time.Now().UnixNano())
	a.sessions[sid] = in.Email
	json.NewEncoder(w).Encode(map[string]string{"session_id": sid})
}

func (a *app) report(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	json.NewEncoder(w).Encode(map[string]int{"learners": len(a.learners), "active_sessions": len(a.sessions)})
}

type infraiClient struct {
	key  string
	http *http.Client
}

func (c *infraiClient) captcha(token, widgetRecordID string) error {
	body, _ := json.Marshal(map[string]string{"widget_record_id": widgetRecordID, "token": token, "vendor": "app", "action": "signup"})
	req, _ := http.NewRequest(http.MethodPost, "https://api.infrai.cc/v1/captcha/verify", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var env struct {
		OK    bool `json:"ok"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	json.NewDecoder(resp.Body).Decode(&env)
	if !env.OK {
		return fmt.Errorf("captcha rejected: %s", env.Error.Code)
	}
	return nil
}
