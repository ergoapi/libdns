package dnspod

import (
	"fmt"
	"strings"

	"github.com/ergoapi/libdns"
	"github.com/ergoapi/util/exstr"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

const (
	TencentCloudDefaultRecordLine = "默认"
)

type Provider struct {
	accessKey    string
	accessSecret string
}

func (p *Provider) Secret(opt libdns.Option) {
	p.accessKey = opt.Key
	p.accessSecret = opt.Secret
}

func (p *Provider) client() *dnspod.Client {
	credential := common.NewCredential(
		p.accessKey,
		p.accessSecret,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
	client, _ := dnspod.NewClient(credential, "", cpf)
	return client
}

func (p *Provider) GetDomainList() ([]libdns.Domain, error) {
	request := dnspod.NewDescribeDomainListRequest()
	request.Offset = common.Int64Ptr(0)
	request.Limit = common.Int64Ptr(3000)
	domainList := make([]libdns.Domain, 0)
	totalCount := int64(100)
	for *request.Offset < totalCount {
		response, err := p.client().DescribeDomainList(request)
		if err != nil {
			return nil, err
		}
		if response.Response.DomainList != nil && len(response.Response.DomainList) > 0 {
			for _, domain := range response.Response.DomainList {
				domainList = append(domainList, libdns.Domain{
					Domain: *domain.Name,
				})
			}
		}
		totalCount = int64(*response.Response.DomainCountInfo.AllTotal)
		request.Offset = common.Int64Ptr(*request.Offset + int64(len(response.Response.DomainList)))
	}
	return domainList, nil
}

func (p *Provider) GetRecordList(domain string) ([]libdns.Record, error) {
	request := dnspod.NewDescribeRecordListRequest()
	request.Domain = common.StringPtr(domain)
	request.Offset = common.Uint64Ptr(0)
	request.Limit = common.Uint64Ptr(3000)
	domainList := make([]libdns.Record, 0)
	totalCount := uint64(100)
	for *request.Offset < totalCount {
		response, err := p.client().DescribeRecordList(request)
		if err != nil {
			return nil, err
		}
		if response.Response.RecordList != nil && len(response.Response.RecordList) > 0 {
			for _, record := range response.Response.RecordList {
				if *record.Name == "@" && *record.Type == "NS" { // Special Record, Skip it.
					continue
				}
				domainList = append(domainList, libdns.Record{
					ID:     fmt.Sprintf("%v", *record.RecordId),
					Value:  *record.Value,
					Name:   *record.Name,
					Type:   *record.Type,
					Status: *record.Status,
					TTL:    int64(*record.TTL),
					Weight: int64(*record.Weight),
				})
			}
		}
		totalCount = *response.Response.RecordCountInfo.TotalCount
		request.Offset = common.Uint64Ptr(*request.Offset + uint64(len(response.Response.RecordList)))
	}
	return domainList, nil
}

func (p *Provider) CreateRecord(domain libdns.Domain, record libdns.Record) error {
	request := dnspod.NewCreateRecordRequest()
	request.Domain = common.StringPtr(domain.Domain)
	request.RecordType = common.StringPtr(record.Type)
	request.Value = common.StringPtr(record.Value)
	request.RecordLine = common.StringPtr(TencentCloudDefaultRecordLine)
	request.SubDomain = common.StringPtr(getSubDomain(record.Name))
	if record.Status == libdns.RecordStatusDISABLE {
		request.Status = common.StringPtr(libdns.RecordStatusDISABLE)
	}
	if record.TTL > 600 {
		request.TTL = common.Uint64Ptr(uint64(record.TTL))
	}
	if _, err := p.client().CreateRecord(request); err != nil {
		return err
	}
	return nil
}

func (p *Provider) DeleteRecord(domain string, recordID string) error {
	request := dnspod.NewDeleteRecordRequest()
	request.Domain = common.StringPtr(domain)
	request.RecordId = common.Uint64Ptr(exstr.Str2UInt64(recordID))
	if _, err := p.client().DeleteRecord(request); err != nil {
		return err
	}
	return nil
}

func (p *Provider) ModifyRecord(domain string, record libdns.Record) error {
	request := dnspod.NewModifyRecordRequest()
	request.Domain = common.StringPtr(domain)
	request.RecordType = common.StringPtr(record.Type)
	request.RecordLine = common.StringPtr(TencentCloudDefaultRecordLine)
	request.Value = common.StringPtr(record.Value)
	request.RecordId = common.Uint64Ptr(exstr.Str2UInt64(record.ID))
	if record.TTL > 600 {
		request.TTL = common.Uint64Ptr(uint64(record.TTL))
	}
	request.Status = common.StringPtr(record.Status)
	if _, err := p.client().ModifyRecord(request); err != nil {
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
	libdns.Register("dnspod", &Provider{})
}
