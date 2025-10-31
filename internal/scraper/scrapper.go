package scraper

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/madhav-poojari/bharat-digital/internal/models"
)

var (
	ColumnApprovedLabourBudget       = "Approved Labour Budget"
	ColumnPersondaysCentralLiability = "Persondays of Central Liability so far"
	ColumnTotalHouseholdsWorked      = "Total Households Worked"
	ColumnTotalJobcardsIssued        = "Total No of JobCards issued"
	ColumnFinYear                    = "fin year"
	ColumnMonth                      = "Month"
	ColumnStateName                  = "State Name"
	ColumnDistrictCode               = "District Code"
	ColumnDistrictName               = "District Name"
)

type Scraper struct {
	client *http.Client
	apiKey string
}

func New(apiKey string, timeout time.Duration) *Scraper {
	return &Scraper{
		client: &http.Client{Timeout: timeout},
		apiKey: apiKey,
	}
}

// fetchCSV returns parsed CSV rows as map header->value
func (s *Scraper) FetchCSV(ctx context.Context, stateName, finYear string) ([]map[string]string, error) {
	url := fmt.Sprintf("https://api.data.gov.in/resource/ee03643a-ee4c-48c2-ac30-9f2ff26ab722?api-key=%s&format=csv&limit=20000&filters%%5Bstate_name%%5D=%s&filters%%5Bfin_year%%5D=%s", s.apiKey, stateName, finYear)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("accept", "application/xml") // matching sample; API returns CSV body though

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("api status %d", resp.StatusCode)
	}

	r := csv.NewReader(resp.Body)
	r.TrimLeadingSpace = true
	headers, err := r.Read()
	if err != nil {
		return nil, err
	}
	// normalize headers
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}

	var rows []map[string]string
	for {
		rec, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		// map row
		m := map[string]string{}
		for i, v := range rec {
			if i < len(headers) {
				m[headers[i]] = strings.TrimSpace(v)
			}
		}
		rows = append(rows, m)
	}
	return rows, nil
}

func parseFloatZero(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "NA") {
		return 0
	}
	s = strings.ReplaceAll(s, ",", "")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// Convert rows into month cache entries and aggregated yearly map
func RowsToCache(rows []map[string]string) (map[string]models.CacheValue, map[string]models.CacheValue) {
	monthPairs := map[string]models.CacheValue{}
	yearAgg := map[string]models.CacheValue{} // keyed by districtcode_FY
	for _, r := range rows {
		dc := r[ColumnDistrictCode]
		dn := r[ColumnDistrictName]
		fy := r[ColumnFinYear]
		month := r[ColumnMonth][:3]

		approved := parseFloatZero(r[ColumnApprovedLabourBudget])
		persondays := parseFloatZero(r[ColumnPersondaysCentralLiability])
		households := parseFloatZero(r[ColumnTotalHouseholdsWorked])
		jobcards := parseFloatZero(r[ColumnTotalJobcardsIssued])

		// month key
		monthKey := fmt.Sprintf("%s_FY%s_%s", dc, fy, month)
		cv := models.CacheValue{
			DistrictCode:               dc,
			DistrictName:               dn,
			FinYear:                    fy,
			Month:                      month,
			Type:                       "month",
			ApprovedLabourBudget:       approved,
			PersondaysCentralLiability: persondays,
			TotalHouseholdsWorked:      households,
			TotalJobcardsIssued:        jobcards,
		}
		monthPairs[monthKey] = cv

		// aggregate
		yearKey := fmt.Sprintf("%s_FY%s", dc, fy)
		agg := yearAgg[yearKey]
		if agg.DistrictCode == "" {
			agg.DistrictCode = dc
			agg.DistrictName = dn
			agg.FinYear = fy
			agg.Type = "year"
		}
		agg.ApprovedLabourBudget += approved
		agg.PersondaysCentralLiability += persondays
		agg.TotalHouseholdsWorked += households
		agg.TotalJobcardsIssued += jobcards
		yearAgg[yearKey] = agg
	}
	return monthPairs, yearAgg
}
