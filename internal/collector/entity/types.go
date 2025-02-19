package entity

type QratorDomainHTTPStats struct {
	Result HTTPStatsResult `json:"result"`
	Error  *string         `json:"error"` // Use pointer to distinguish between null and empty string
	ID     int             `json:"id"`
}

type HTTPStatsResult struct {
	Time      int64              `json:"time"`
	Requests  float64            `json:"requests"`
	Responses HTTPStatsDurations `json:"responses"`
	Errors    HTTPStatErrors     `json:"errors"`
}

type HTTPStatsDurations struct {
	Duration0000_0200 float64 `json:"0000_0200"`
	Duration0200_0500 float64 `json:"0200_0500"`
	Duration0500_0700 float64 `json:"0500_0700"`
	Duration0700_1000 float64 `json:"0700_1000"`
	Duration1000_1500 float64 `json:"1000_1500"`
	Duration1500_2000 float64 `json:"1500_2000"`
	Duration2000_5000 float64 `json:"2000_5000"`
	Duration5000_Inf  float64 `json:"5000_inf"`
}

type HTTPStatErrors struct {
	Total   float64 `json:"total"`
	Code500 float64 `json:"500"`
	Code501 float64 `json:"501"`
	Code502 float64 `json:"502"`
	Code503 float64 `json:"503"`
	Code504 float64 `json:"504"`
	Code4xx float64 `json:"4xx"`
}

type QratorDomainBillStats struct {
	Result float64 `json:"result"`
	Error  *string `json:"error"` // Use pointer to distinguish between null and empty string
	ID     int     `json:"id"`
}

type QratorDomainIPStats struct {
	Result DomainIPStatsResult
	Error  *string
	ID     int
}

type DomainIPStatsResult struct {
	Time      int64         `json:"time"`
	Bandwidth IPStatistics  `json:"bandwidth"`
	Packets   IPStatistics  `json:"packets"`
	Blacklist BlacklistStat `json:"blacklist"`
}

type IPStatistics struct {
	Input  float64 `json:"input"`
	Passed float64 `json:"passed"`
	Output float64 `json:"output"`
}

type BlacklistStat struct {
	Qrator float64 `json:"qrator"`
	API    float64 `json:"api"`
	WAF    float64 `json:"waf"`
	Custom float64 `json:"custom"`
}

type QratorResponseDomainName struct {
	Result string `json:"result"`
	Error  string `json:"error"`
	ID     int    `json:"id"`
}

type QratorRequest struct {
	Method string `json:"method"`
	Params string `json:"params"`
	ID     int    `json:"id"`
}

type QratorDomain struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	QratorIP string `json:"qratorIp"`
}

type QratorDomains struct {
	Domains []QratorDomain `json:"result"`
	Error   string         `json:"error"`
	ID      int            `json:"id"`
}

type QratorPing struct {
	Result []string `json:"result"`
	Error  string   `json:"error"`
	ID     int      `json:"id"`
}
