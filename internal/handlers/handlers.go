
package handlers
import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/madhav-poojari/bharat-digital/internal/cache"
	"github.com/madhav-poojari/bharat-digital/internal/config"
	"github.com/madhav-poojari/bharat-digital/internal/cron"
	"github.com/madhav-poojari/bharat-digital/internal/districts"
	"github.com/madhav-poojari/bharat-digital/internal/models"

	"github.com/gorilla/mux"
)

type App struct {
	Cache  *cache.Client
	Croner *cron.Croner
	CFG    *config.Config
}

type apiResp struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (a *App) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/health", a.health).Methods("GET")
	r.HandleFunc("/cron/trigger", a.triggerCron).Methods("POST")
	r.HandleFunc("/state/all", a.getStateAll).Methods("GET")
	r.HandleFunc("/district/{districtcode}", a.getDistrict).Methods("GET")
}

func (a *App) writeJSON(w http.ResponseWriter, status int, success bool, msg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiResp{Success: success, Message: msg, Data: data})
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	a.writeJSON(w, 200, true, "ok", map[string]string{})
}

func (a *App) triggerCron(w http.ResponseWriter, r *http.Request) {
	go a.Croner.RunOnce(context.Background())
	a.writeJSON(w, 200, true, "cron started", nil)
}

// GET /state/all?state_name=maharashtra&fy=2024-2025
func (a *App) getStateAll(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state_name")
	if state == "" {
		a.writeJSON(w, 400, false, "state_name required", nil)
		return
	}
	fy := r.URL.Query().Get("fy")
	if fy == "" {
		if len(a.CFG.FYList) > 0 {
			fy = a.CFG.FYList[0] // newest FY by config
		} else {
			a.writeJSON(w, 400, false, "fy required", nil)
			return
		}
	}

	// only MAHARASHTRA allowed in MVP
	if !strings.EqualFold(state, "maharashtra") {
		a.writeJSON(w, 400, false, "only maharashtra supported in MVP", nil)
		return
	}

	// prepare keys for all MH districts (year keys)
	keys := make([]string, 0, len(districts.MHOrdered))
	for _, dc := range districts.MHOrdered {
		keys = append(keys, fmt.Sprintf("%s_FY%s", dc, fy))
	}

	rows, err := a.Cache.MGetJSON(r.Context(), keys...)
	if err != nil {
		log.Printf("redis mget error: %v", err)
		a.writeJSON(w, 500, false, "internal error", nil)
		return
	}

	// build result list: if a row is nil we skip it (per your instruction we return empty list if none found)
	out := []models.CacheValue{}
	for _, cv := range rows {
		if cv != nil {
			out = append(out, *cv)
		}
	}
	if len(out) == 0 {
		a.writeJSON(w, 200, true, "no rows found", []models.CacheValue{})
		return
	}
	a.writeJSON(w, 200, true, "ok", out)
}

// GET /district/{districtcode}?type=month|year&startyear=2022&endyear=2024
func (a *App) getDistrict(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dc := vars["districtcode"]
	typ := r.URL.Query().Get("type")
	if typ == "" {
		typ = "year"
	}
	startStr := r.URL.Query().Get("startyear")
	endStr := r.URL.Query().Get("endyear")
	if startStr == "" || endStr == "" {
		a.writeJSON(w, 400, false, "startyear and endyear required (integers)", nil)
		return
	}
	start, err1 := strconv.Atoi(startStr)
	end, err2 := strconv.Atoi(endStr)
	if err1 != nil || err2 != nil || start > end {
		a.writeJSON(w, 400, false, "invalid startyear/endyear", nil)
		return
	}

	// build list of FY strings from start..end (inclusive) as "2022-2023"
	fys := make([]string, 0, end-start+1)
	for y := start; y < end; y++ {
		fys = append(fys, fmt.Sprintf("%d-%d", y, y+1))
	}

	keys := []string{}
	if typ == "year" {
		for _, fy := range fys {
			keys = append(keys, fmt.Sprintf("%s_FY%s", dc, fy))
		}
	} else { // month
		months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
		for _, fy := range fys {
			for _, m := range months {
				keys = append(keys, fmt.Sprintf("%s_FY%s_%s", dc, fy, m))
			}
		}
	}

	rows, err := a.Cache.MGetJSON(r.Context(), keys...)
	if err != nil {
		log.Printf("redis mget error: %v", err)
		a.writeJSON(w, 500, false, "internal error", nil)
		return
	}

	// ordered result - return non-nil entries in order
	out := []models.CacheValue{}
	for _, cv := range rows {
		if cv != nil {
			out = append(out, *cv)
		}
	}
	if len(out) == 0 {
		a.writeJSON(w, 200, true, "no rows found", []models.CacheValue{})
		return
	}
	a.writeJSON(w, 200, true, "ok", out)
}
