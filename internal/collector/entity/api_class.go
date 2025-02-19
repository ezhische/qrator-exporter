package entity

type APIMethod string

const (
	HTTP       APIMethod = "statistics_current_http"
	Bill       APIMethod = "statistics_billable"
	IP         APIMethod = "statistics_current_ip"
	GetDomains APIMethod = "domains_get"
	Ping       APIMethod = "source_ips_get"
	Name       APIMethod = "name_get"
)

func (c APIMethod) String() string {
	return string(c)
}

type MethodClass string

const (
	Client MethodClass = "client"
	Domain MethodClass = "domain"
)

func (c MethodClass) String() string {
	return string(c)
}
