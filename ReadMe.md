# mgnrega-mvp

Env vars:
- REDIS_ADDR (default 127.0.0.1:6379)
- REDIS_PASS (optional)
- DATA_GOV_API_KEY (required)
- PORT (default 8080)
- CRON_SCHEDULE (default "0 2 * * *")
- FY_LIST (optional; comma-separated; e.g. "2024-2025,2023-2024,..."). If omitted, last 6 FYs starting from current year will be used.

Run:
1. Start Redis.
2. go mod tidy
3. go run ./cmd/api

Endpoints:
- GET /health
- POST /cron/trigger        -> manual trigger (starts in background)
- GET /state/all?state_name=MAHARASHTRA&fy=2024-2025
- GET /district/{districtcode}?type=month|year&startyear=2022&endyear=2024

Response envelope:
{ "success": bool, "message": string, "data": ... }

Notes:
- Cron pauses 30s between each FY API call.
- Uses Redis MSET / MGET for bulk IO.
- Missing/NA numeric values treated as 0.
- Keys:
  - Monthly: <districtcode>_FY<fin_year>_<Month> (e.g. 1806_FY2024-2025_Feb)
  - Yearly:  <districtcode>_FY<fin_year> (e.g. 1806_FY2024-2025)
