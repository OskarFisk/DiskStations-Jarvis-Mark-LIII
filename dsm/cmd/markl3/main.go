package main

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

type Config struct { GeminiAPIKey string `json:"gemini_api_key"`; Model string `json:"model"` }

type App struct { dataDir string; static http.Handler }

func main() {
    dataDir := os.Getenv("MARK_DATA_DIR")
    if dataDir == "" { dataDir = "/var/packages/DiskStations-Jarvis-Mark-LIII/var" }
    _ = os.MkdirAll(dataDir, 0700)
    app := &App{dataDir: dataDir, static: http.FileServer(http.Dir("/var/packages/DiskStations-Jarvis-Mark-LIII/target/www"))}
    mux := http.NewServeMux()
    mux.HandleFunc("/api/health", app.health)
    mux.HandleFunc("/api/config", app.config)
    mux.HandleFunc("/api/chat", app.chat)
    mux.HandleFunc("/", app.index)
    addr := "0.0.0.0:8080"
    log.Printf("MARK LIII listening on %s", addr)
    log.Fatal(http.ListenAndServe(addr, mux))
}

func (a *App) index(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" { http.NotFound(w, r); return }
    a.static.ServeHTTP(w, r)
}
func (a *App) health(w http.ResponseWriter, r *http.Request) { writeJSON(w, map[string]any{"ok":true,"service":"MARK LIII","platform":"Synology DS413j / 88f628x","version":"1.0.0"}) }
func (a *App) configPath() string { return filepath.Join(a.dataDir, "config.json") }
func (a *App) loadConfig() Config {
    b, err := os.ReadFile(a.configPath()); if err != nil { return Config{Model:"gemini-2.5-flash"} }
    var c Config; if json.Unmarshal(b,&c)!=nil { c=Config{} }; if c.Model=="" { c.Model="gemini-2.5-flash" }; return c
}
func (a *App) config(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" { c:=a.loadConfig(); writeJSON(w,map[string]any{"configured":c.GeminiAPIKey!="","model":c.Model}); return }
    if r.Method != "POST" { http.Error(w,"method not allowed",405); return }
    var in struct{ GeminiAPIKey string `json:"gemini_api_key"`; Model string `json:"model"` }
    if err:=json.NewDecoder(r.Body).Decode(&in); err!=nil { http.Error(w,"invalid json",400); return }
    if in.Model=="" { in.Model="gemini-2.5-flash" }
    b,_:=json.MarshalIndent(Config{GeminiAPIKey:strings.TrimSpace(in.GeminiAPIKey),Model:in.Model},"","  ")
    if err:=os.WriteFile(a.configPath(),b,0600); err!=nil { http.Error(w,"cannot save config",500); return }
    writeJSON(w,map[string]any{"ok":true})
}
func (a *App) chat(w http.ResponseWriter, r *http.Request) {
    if r.Method!="POST" { http.Error(w,"method not allowed",405); return }
    var in struct{ Message string `json:"message"` }
    if err:=json.NewDecoder(r.Body).Decode(&in); err!=nil || strings.TrimSpace(in.Message)=="" { http.Error(w,"message required",400); return }
    c:=a.loadConfig(); if c.GeminiAPIKey=="" { http.Error(w,"Gemini API key is not configured",409); return }
    out,err:=gemini(c.GeminiAPIKey,c.Model,in.Message); if err!=nil { log.Printf("gemini: %v",err); http.Error(w,"Gemini request failed",502); return }
    writeJSON(w,map[string]any{"reply":out})
}
func gemini(key, model, message string) (string,error) {
    body:=fmt.Sprintf(`{"contents":[{"parts":[{"text":%s}]}]}`,quote(message))
    req,err:=http.NewRequest("POST","https://generativelanguage.googleapis.com/v1beta/models/"+model+":generateContent?key="+key,strings.NewReader(body)); if err!=nil{return "",err}
    req.Header.Set("Content-Type","application/json")
    resp,err:=http.DefaultClient.Do(req); if err!=nil{return "",err}; defer resp.Body.Close(); b,_:=io.ReadAll(resp.Body); if resp.StatusCode/100!=2{return "",fmt.Errorf("HTTP %s: %s",resp.Status,string(b))}
    var v struct{ Candidates []struct{ Content struct{ Parts []struct{ Text string `json:"text"` } `json:"parts"` } `json:"content"` } `json:"candidates"` }
    if err:=json.Unmarshal(b,&v); err!=nil{return "",err}; if len(v.Candidates)==0 || len(v.Candidates[0].Content.Parts)==0{return "",fmt.Errorf("empty response")}; return v.Candidates[0].Content.Parts[0].Text,nil
}
func quote(s string) string { b,_:=json.Marshal(s); return string(b) }
func writeJSON(w http.ResponseWriter,v any){ w.Header().Set("Content-Type","application/json"); _=json.NewEncoder(w).Encode(v) }
