package main

import (
    "embed"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

//go:embed www/index.html
var webFS embed.FS

type Config struct {
    GeminiAPIKey string `json:"gemini_api_key"`
    Model string `json:"model"`
}

type App struct { dataDir string }

func main() {
    dataDir := os.Getenv("MARK_DATA_DIR")
    if dataDir == "" { dataDir = "/var/packages/DiskStations-Jarvis-Mark-LIII/var" }
    if err := os.MkdirAll(dataDir, 0700); err != nil { log.Fatal(err) }
    app := &App{dataDir: dataDir}
    mux := http.NewServeMux()
    mux.HandleFunc("/api/health", app.health)
    mux.HandleFunc("/api/config", app.config)
    mux.HandleFunc("/api/chat", app.chat)
    mux.HandleFunc("/", app.index)
    log.Printf("MARK LIII listening on 0.0.0.0:8080")
    log.Fatal(http.ListenAndServe("0.0.0.0:8080", mux))
}

func (a *App) index(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" { http.NotFound(w, r); return }
    data, err := webFS.ReadFile("www/index.html")
    if err != nil { http.Error(w, "MARK LIII UI unavailable", 500); return }
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    _, _ = w.Write(data)
}
func (a *App) health(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, map[string]any{"ok":true,"service":"MARK LIII","platform":"Synology DS413j / 88f628x","version":"1.0.1"})
}
func (a *App) configPath() string { return filepath.Join(a.dataDir, "config.json") }
func (a *App) loadConfig() Config {
    b, err := os.ReadFile(a.configPath()); if err != nil { return Config{Model:"gemini-2.5-flash"} }
    var c Config; if json.Unmarshal(b,&c)!=nil { c=Config{} }; if c.Model=="" { c.Model="gemini-2.5-flash" }; return c
}
func (a *App) config(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet { c:=a.loadConfig(); writeJSON(w,map[string]any{"configured":c.GeminiAPIKey!="","model":c.Model}); return }
    if r.Method != http.MethodPost { http.Error(w,"method not allowed",405); return }
    var in struct{ GeminiAPIKey string `json:"gemini_api_key"`; Model string `json:"model"` }
    if err:=json.NewDecoder(r.Body).Decode(&in); err!=nil { http.Error(w,"invalid json",400); return }
    if in.Model=="" { in.Model="gemini-2.5-flash" }
    b,err:=json.MarshalIndent(Config{GeminiAPIKey:strings.TrimSpace(in.GeminiAPIKey),Model:in.Model},"","  "); if err!=nil { http.Error(w,"cannot encode config",500); return }
    if err:=os.WriteFile(a.configPath(),b,0600); err!=nil { http.Error(w,"cannot save config",500); return }
    writeJSON(w,map[string]any{"ok":true})
}
func (a *App) chat(w http.ResponseWriter, r *http.Request) {
    if r.Method!=http.MethodPost { http.Error(w,"method not allowed",405); return }
    var in struct{ Message string `json:"message"` }
    if err:=json.NewDecoder(r.Body).Decode(&in); err!=nil || strings.TrimSpace(in.Message)=="" { http.Error(w,"message required",400); return }
    c:=a.loadConfig(); if c.GeminiAPIKey=="" { http.Error(w,"Gemini API key is not configured",409); return }
    out,err:=gemini(c.GeminiAPIKey,c.Model,in.Message); if err!=nil { log.Printf("gemini: %v",err); http.Error(w,"Gemini request failed",502); return }
    writeJSON(w,map[string]any{"reply":out})
}
func gemini(key, model, message string) (string,error) {
    payload:=map[string]any{"contents":[]any{map[string]any{"parts":[]any{map[string]string{"text":message}}}}}
    body,err:=json.Marshal(payload); if err!=nil{return "",err}
    endpoint:="https://generativelanguage.googleapis.com/v1beta/models/"+model+":generateContent?key="+key
    req,err:=http.NewRequest(http.MethodPost,endpoint,strings.NewReader(string(body))); if err!=nil{return "",err}; req.Header.Set("Content-Type","application/json")
    resp,err:=http.DefaultClient.Do(req); if err!=nil{return "",err}; defer resp.Body.Close(); data,_:=io.ReadAll(resp.Body)
    if resp.StatusCode/100!=2{return "",fmt.Errorf("HTTP %s: %s",resp.Status,string(data))}
    var result struct{ Candidates []struct{ Content struct{ Parts []struct{ Text string `json:"text"` } `json:"parts"` } `json:"content"` } `json:"candidates"` }
    if err:=json.Unmarshal(data,&result); err!=nil{return "",err}; if len(result.Candidates)==0 || len(result.Candidates[0].Content.Parts)==0{return "",fmt.Errorf("empty response")}; return result.Candidates[0].Content.Parts[0].Text,nil
}
func writeJSON(w http.ResponseWriter,v any){ w.Header().Set("Content-Type","application/json"); _=json.NewEncoder(w).Encode(v) }
