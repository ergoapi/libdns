package libdns

import "fmt"

const (
	DNSEmptyPrefix      = "@"
	RecordStatusDefault = "ENABLE"
	RecordStatusDISABLE = "DISABLE"
)

type Domain struct {
	Domain   string
	Provider string
}

type Option struct {
	Key    string
	Secret string
}

type Record struct {
	ID     string `json:"id,omitempty" name:"id"`         // 记录ID
	Value  string `json:"value,omitempty" name:"value"`   // 记录值
	Name   string `json:"Name,omitempty" name:"Name"`     // 主机名
	Type   string `json:"Type,omitempty" name:"Type"`     // 记录类型
	Status string `json:"status,omitempty" name:"status"` // 记录状态，启用：ENABLE，暂停：DISABLE
	TTL    int64  `json:"TTL,omitempty" name:"TTL"`       // 记录缓存时间
	Weight int64  `json:"Weight,omitempty" name:"Weight"` // 记录权重，
}

type Provider interface {
	GetDomainList() ([]Domain, error)
	GetRecordList(domain string) ([]Record, error)
	CreateRecord(domain Domain, record Record) error
	DeleteRecord(domain string, recordID string) error
	ModifyRecord(domain string, record Record) error
	Secret(opt Option)
}

func NewDns(name string, opt Option) (Provider, error) {
	adapter, ok := adapters[name]
	if !ok {
		return nil, fmt.Errorf("dns: unknown adapter '%s'", name)
	}
	adapter.Secret(opt)
	return adapter, nil
}

var adapters = make(map[string]Provider)

func Register(name string, adapter Provider) {
	if adapter == nil {
		panic("dns: cannot register adapter with nil value")
	}
	if _, dup := adapters[name]; dup {
		panic(fmt.Errorf("dns: cannot register adapter '%s' twice", name))
	}
	adapters[name] = adapter
}
