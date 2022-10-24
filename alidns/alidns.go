package alidns

import (
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	alidns "github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/ergoapi/libdns"
)

type Provider struct {
	accessKey    string
	accessSecret string
}

func (p *Provider) Secret(opt libdns.Option) {
	p.accessKey = opt.Key
	p.accessSecret = opt.Secret
}

func (p *Provider) client() *alidns.Client {
	config := sdk.NewConfig()
	credential := credentials.NewAccessKeyCredential(p.accessKey, p.accessSecret)
	client, _ := alidns.NewClientWithOptions("cn-hangzhou", config, credential)
	return client
}

func (p *Provider) GetDomainList() ([]libdns.Domain, error) {
	request := alidns.CreateDescribeDomainsRequest()
	request.Scheme = "https"
	pageNum := 1
	pageSize := 100
	request.PageSize = requests.NewInteger(pageSize)
	domainList := make([]libdns.Domain, 0)
	totalCount := int64(100)
	for int64(pageNum-1)*int64(pageSize) <= totalCount {
		request.PageNumber = requests.NewInteger(pageNum)
		response, err := p.client().DescribeDomains(request)
		if err != nil {
			return nil, err
		}
		if len(response.Domains.Domain) > 0 {
			for _, domain := range response.Domains.Domain {
				domainList = append(domainList, libdns.Domain{
					Domain: domain.DomainName,
				})
			}
		}
		totalCount = response.TotalCount
		pageNum++
	}
	return domainList, nil
}

func (p *Provider) GetRecordList(domain string) ([]libdns.Record, error) {
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.Scheme = "https"
	pageNum := 1
	pageSize := 100
	request.DomainName = domain
	request.PageSize = requests.NewInteger(pageSize)
	domainList := make([]libdns.Record, 0)
	totalCount := int64(100)
	for int64(pageNum-1)*int64(pageSize) < totalCount {
		request.PageNumber = requests.NewInteger(pageNum)
		response, err := p.client().DescribeDomainRecords(request)
		if err != nil {
			return nil, err
		}
		if len(response.DomainRecords.Record) > 0 {
			for _, record := range response.DomainRecords.Record {
				if record.RR == "@" && record.Type == "NS" { // Special Record, Skip it.
					continue
				}
				domainList = append(domainList, libdns.Record{
					ID:     record.RecordId,
					Value:  record.Value,
					Name:   record.RR,
					Type:   record.Type,
					Status: record.Status,
					TTL:    record.TTL,
					Weight: int64(record.Weight),
				})
			}
		}
		totalCount = response.TotalCount
		pageNum++
	}
	return domainList, nil
}

func (p *Provider) CreateRecord(domain libdns.Domain, record libdns.Record) error {
	request := alidns.CreateAddDomainRecordRequest()
	request.Scheme = "https"
	request.DomainName = domain.Domain
	request.Type = record.Type
	request.Value = record.Value
	request.RR = getSubDomain(record.Name)
	if record.TTL > 600 {
		request.TTL = requests.NewInteger(int(record.TTL))
	}
	if _, err := p.client().AddDomainRecord(request); err != nil {
		return err
	}
	return nil
}

func (p *Provider) DeleteRecord(domain string, recordID string) error {
	request := alidns.CreateDeleteDomainRecordRequest()
	request.Scheme = "https"
	request.RecordId = recordID
	if _, err := p.client().DeleteDomainRecord(request); err != nil {
		return err
	}
	return nil
}

func (p *Provider) ModifyRecord(domain string, record libdns.Record) error {
	if record.Status != "" {
		request := alidns.CreateSetDomainRecordStatusRequest()
		request.Scheme = "https"
		request.Status = record.Status
		request.RecordId = record.ID
		if _, err := p.client().SetDomainRecordStatus(request); err != nil {
			return err
		}
		return nil
	}
	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"
	request.RR = record.Name
	request.Type = record.Type
	request.Value = record.Value
	request.RecordId = record.ID
	if record.TTL > 600 {
		request.TTL = requests.NewInteger(int(record.TTL))
	}
	if _, err := p.client().UpdateDomainRecord(request); err != nil {
		return err
	}
	return nil
}

func getSubDomain(name string) string {
	name = strings.TrimSuffix(name, ".")
	if name == "" {
		return libdns.DNSEmptyPrefix
	}
	return name
}

func init() {
	libdns.Register("alidns", &Provider{})
}
