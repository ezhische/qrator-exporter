package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ezhische/qrator-exporter/internal/collector/entity"
	log "github.com/sirupsen/logrus"
)

func (c *Collector) qratorPostRequest(methodClass entity.MethodClass, id int, method entity.APIMethod) (*http.Response, error) {
	reqURL := fmt.Sprintf("%s/%s/%d", c.config.qratorAPIURL, methodClass.String(), id)
	reqBody := entity.QratorRequest{
		Method: method.String(),
		ID:     1,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("cannot create new request: %w", err)
	}
	defer req.Body.Close()
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Qrator-Auth", c.config.aPIKey)

	client := c.client
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making new request: %w", err)
	}
	return response, nil
}

func (c *Collector) getQratorDomains() ([]entity.QratorDomain, error) {
	if len(c.config.domainsList) > 0 {
		var list []entity.QratorDomain
		for _, domain := range c.config.domainsList {
			qds, err := c.getQratorDomainName(domain)
			if err != nil {
				log.Errorf("got error while getting domain name for id: %v %v", domain, err)
				continue
			}
			list = append(list, entity.QratorDomain{ID: domain, Name: qds.Result})
		}
		return list, nil
	}

	r, err := c.qratorPostRequest(entity.Client, c.config.clientID, entity.GetDomains)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	qds := entity.QratorDomains{}
	err = json.NewDecoder(r.Body).Decode(&qds)
	if err != nil {
		log.Errorf("Can't decode received json: %v", err)
		return nil, err
	}
	if qds.Error != "" {
		log.Errorf("wrong request: %s", qds.Error)
		return nil, fmt.Errorf("wrong request: %s", qds.Error)
	}
	return qds.Domains, nil
}

func (c *Collector) getQratorDomainName(domainID int) (*entity.QratorResponseDomainName, error) {
	r, err := c.qratorPostRequest(entity.Domain, domainID, entity.Name)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	qds := &entity.QratorResponseDomainName{}

	err = json.NewDecoder(r.Body).Decode(&qds)
	if err != nil {
		log.Errorf("Can't decode received json: %v", err)
		return nil, fmt.Errorf("parse error %w", err)
	}
	return qds, nil
}

func (c *Collector) qratorCheck() error {
	r, err := c.qratorPostRequest(entity.Client, c.config.clientID, entity.Ping)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	ping := entity.QratorPing{}
	err = json.NewDecoder(r.Body).Decode(&ping)
	if err != nil {
		return fmt.Errorf("got error while decoding json. %w", err)
	}
	if ping.Error != "" {
		return fmt.Errorf("got error in response: %s", ping.Error)
	}
	return nil
}

func (c *Collector) getQratorDomainHTTPStats(qd entity.QratorDomain) (*entity.QratorDomainHTTPStats, error) {
	r, err := c.qratorPostRequest("domain", qd.ID, "statistics_current_http")
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	stats := &entity.QratorDomainHTTPStats{}
	err = json.NewDecoder(r.Body).Decode(&stats)
	if err != nil {
		return nil, fmt.Errorf("parse error for domain %s:%w", qd.Name, err)
	}
	if stats.Error != nil {
		return nil, fmt.Errorf("wrong request for domain %s : %s", qd.Name, *stats.Error)
	}
	return stats, nil
}

func (c *Collector) getQratorDomainIPStats(qd entity.QratorDomain) (*entity.QratorDomainIPStats, error) {
	r, err := c.qratorPostRequest(entity.Domain, qd.ID, entity.IP)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	stats := &entity.QratorDomainIPStats{}
	err = json.NewDecoder(r.Body).Decode(&stats)
	if err != nil {
		return nil, fmt.Errorf("parse error for domain %s:%w", qd.Name, err)
	}
	if stats.Error != nil {
		return nil, fmt.Errorf("wrong request for domain %s : %s", qd.Name, *stats.Error)
	}
	return stats, nil
}

func (c *Collector) getQratorDomainBillableStats(qd entity.QratorDomain) (*entity.QratorDomainBillStats, error) {
	r, err := c.qratorPostRequest(entity.Domain, qd.ID, entity.Bill)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	stats := &entity.QratorDomainBillStats{}
	err = json.NewDecoder(r.Body).Decode(&stats)
	if err != nil {
		return nil, fmt.Errorf("parse error for domain %s:%w", qd.Name, err)
	}
	if stats.Error != nil {
		return nil, fmt.Errorf("wrong request for domain %s : %s", qd.Name, *stats.Error)
	}
	return stats, nil
}
