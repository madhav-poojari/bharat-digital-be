package models

type CacheValue struct {
	DistrictCode               string  `json:"district_code"`
	DistrictName               string  `json:"district_name"`
	FinYear                    string  `json:"fin_year"`
	Month                      string  `json:"month,omitempty"`
	Type                       string  `json:"type"` // "month" or "year"
	ApprovedLabourBudget       float64 `json:"approved_labour_budget"`
	PersondaysCentralLiability float64 `json:"persondays_central_liability"`
	TotalHouseholdsWorked      float64 `json:"total_households_worked"`
	TotalJobcardsIssued        float64 `json:"total_jobcards_issued"`
}
