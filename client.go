package dreamhost

import (
	"strings"

	"github.com/adamantal/go-dreamhost/api"
	"github.com/libdns/libdns"
)

func (p *Provider) init() error {
	client, err := api.NewClient(p.APIKey, nil)
	if err != nil {
		return err
	}
	p.client = *client
	return nil
}

// Custom Record implementation for the Dreamhost provider
type DreamhostRecord struct {
	TypeValue  string
	ValueValue string
	NameValue  string
	rrCache    libdns.RR
}

// RR implements the libdns.Record interface
func (r DreamhostRecord) RR() libdns.RR {
	return r.rrCache
}

func recordFromApiDnsRecord(apiDnsRecord api.DNSRecord) libdns.Record {
	// Create our custom record implementation
	record := &DreamhostRecord{
		TypeValue:  string(apiDnsRecord.Type),
		ValueValue: apiDnsRecord.Value,
		NameValue:  libdns.RelativeName(apiDnsRecord.Record, apiDnsRecord.Zone),
	}

	// Initialize the RR cache
	record.rrCache = libdns.RR{
		Name: record.NameValue,
	}

	return record
}

func apiDnsRecordInputFromRecord(record libdns.Record, zone string) api.DNSRecordInput {
	var recordInput api.DNSRecordInput

	// Try to cast to our custom record type
	dhRecord, ok := record.(*DreamhostRecord)
	if ok {
		recordInput.Type = api.RecordType(dhRecord.TypeValue)
		recordInput.Value = dhRecord.ValueValue
		// Dreamhost expects the record name to be absolute, without a dot at the end
		zone = strings.TrimRight(zone, ".")
		recordInput.Record = libdns.AbsoluteName(dhRecord.NameValue, zone)
	} else {
		// Fallback for non-dreamhost records (probably won't work well)
		rr := record.RR()
		recordInput.Type = api.RecordType("TXT") // Default to TXT
		recordInput.Value = ""                   // Can't extract value
		zone = strings.TrimRight(zone, ".")
		recordInput.Record = libdns.AbsoluteName(rr.Name, zone)
	}

	return recordInput
}
