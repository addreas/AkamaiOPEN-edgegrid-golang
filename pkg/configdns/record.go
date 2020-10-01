package dns

import (
	"context"
	"fmt"
	"net/http"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v2/pkg/session"
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"net"
	"sync"
)

// The record types implemented and their fields are as defined here
// https://developer.akamai.com/api/luna/config-dns/data.html

// Records contains operations available on a Record resource
// See: https://developer.akamai.com/api/cloud_security/edge_dns_zone_management/v2.html
type Records interface {
	// RecordToMap returns a map containing record content
	RecordToMap(context.Context, *RecordBody) map[string]interface{}
	// Return bare bones tsig key struct
	NewRecordBody(context.Context, RecordBody) *RecordBody
	//  GetRecordList retrieves recordset list based on type
	// See: https://developer.akamai.com/api/cloud_security/edge_dns_zone_management/v2.html#getzonerecordsets
	GetRecordList(context.Context, string, string, string) (*RecordSetResponse, error)
	// GetRdata retrieves record rdata, e.g. target
	GetRdata(context.Context, string, string, string) ([]string, error)
	// ProcessRdata
	ProcessRdata(context.Context, []string, string) []string
	// ParseRData parses rdata. returning map
	ParseRData(context.Context, string, []string) map[string]interface{}
	// GetRecord retrieves a recordset and returns as RecordBody
	// See:  https://developer.akamai.com/api/cloud_security/edge_dns_zone_management/v2.html#getzonerecordset
	GetRecord(context.Context, string, string, string) (*RecordBody, error)
	// CreateRecord creates recordset
	// See: https://developer.akamai.com/api/cloud_security/edge_dns_zone_management/v2.html#postzonerecordset
	CreateRecord(context.Context, *RecordBody, string, ...bool) error
	// DeleteRecord removes recordset
	// See: https://developer.akamai.com/api/cloud_security/edge_dns_zone_management/v2.html#deletezonerecordset
	DeleteRecord(context.Context, *RecordBody, string, ...bool) error
	// UpdateRecord replaces the recordset
	// See: https://developer.akamai.com/api/cloud_security/edge_dns_zone_management/v2.html#putzonerecordset
	UpdateRecord(context.Context, *RecordBody, string, ...bool) error
	// FullIPv6 is utility method to convert IP to string
	FullIPv6(context.Context, net.IP) string
	// PadCoordinates is utility method to convert IP to normalize coordinates
	PadCoordinates(context.Context, string) string
}

type RecordBody struct {
	Name       string `json:"name,omitempty"`
	RecordType string `json:"type,omitempty"`
	TTL        int    `json:"ttl,omitempty"`
	// Active field no longer used in v2
	Active bool     `json:"active,omitempty"`
	Target []string `json:"rdata,omitempty"`
	/*
		// Remaining Fields are not used in the v2 API
		Subtype             int    `json:"subtype,omitempty"`                //AfsdbRecord
		Flags               int    `json:"flags,omitempty"`                  //DnskeyRecord Nsec3paramRecord
		Protocol            int    `json:"protocol,omitempty"`               //DnskeyRecord
		Algorithm           int    `json:"algorithm,omitempty"`              //DnskeyRecord DsRecord Nsec3paramRecord RrsigRecord SshfpRecord
		Key                 string `json:"key,omitempty"`                    //DnskeyRecord
		Keytag              int    `json:"keytag,omitempty"`                 //DsRecord RrsigRecord
		DigestType          int    `json:"digest_type,omitempty"`            //DsRecord
		Digest              string `json:"digest,omitempty"`                 //DsRecord
		Hardware            string `json:"hardware,omitempty"`               //HinfoRecord
		Software            string `json:"software,omitempty"`               //HinfoRecord
		Priority            int    `json:"priority,omitempty"`               //MxRecord SrvRecord
		Order               uint16 `json:"order,omitempty"`                  //NaptrRecord
		Preference          uint16 `json:"preference,omitempty"`             //NaptrRecord
		FlagsNaptr          string `json:"flags,omitempty"`                  //NaptrRecord
		Service             string `json:"service,omitempty"`                //NaptrRecord
		Regexp              string `json:"regexp,omitempty"`                 //NaptrRecord
		Replacement         string `json:"replacement,omitempty"`            //NaptrRecord
		Iterations          int    `json:"iterations,omitempty"`             //Nsec3Record Nsec3paramRecord
		Salt                string `json:"salt,omitempty"`                   //Nsec3Record Nsec3paramRecord
		NextHashedOwnerName string `json:"next_hashed_owner_name,omitempty"` //Nsec3Record
		TypeBitmaps         string `json:"type_bitmaps,omitempty"`           //Nsec3Record
		Mailbox             string `json:"mailbox,omitempty"`                //RpRecord
		Txt                 string `json:"txt,omitempty"`                    //RpRecord
		TypeCovered         string `json:"type_covered,omitempty"`           //RrsigRecord
		OriginalTTL         int    `json:"original_ttl,omitempty"`           //RrsigRecord
		Expiration          string `json:"expiration,omitempty"`             //RrsigRecord
		Inception           string `json:"inception,omitempty"`              //RrsigRecord
		Signer              string `json:"signer,omitempty"`                 //RrsigRecord
		Signature           string `json:"signature,omitempty"`              //RrsigRecord
		Labels              int    `json:"labels,omitempty"`                 //RrsigRecord
		Weight              uint16 `json:"weight,omitempty"`                 //SrvRecord
		Port                uint16 `json:"port,omitempty"`                   //SrvRecord
		FingerprintType     int    `json:"fingerprint_type,omitempty"`       //SshfpRecord
		Fingerprint         string `json:"fingerprint,omitempty"`            //SshfpRecord
		PriorityIncrement   int    `json:"priority_increment,omitempty"`     //MX priority Increment
	*/
}

var (
	zoneRecordWriteLock sync.Mutex
)

// Validate validates RecordBody
func (rec *RecordBody) Validate() error {
	return validation.Errors{
		"Name":       validation.Validate(rec.Name, validation.Required),
		"RecordType": validation.Validate(rec.RecordType, validation.Required),
		"TTL":        validation.Validate(rec.TTL, validation.Required),
		"Target":     validation.Validate(rec.Target, validation.Required),
	}.Filter()
}

func (p *dns) RecordToMap(ctx context.Context, record *RecordBody) map[string]interface{} {

	logger := p.Log(ctx)
	logger.Debug("RecordToMap")

	if err := record.Validate; err != nil {
		logger.Errorf("Record to map failed. %w", err)
		return nil
	}

	return map[string]interface{}{
		"name":       record.Name,
		"ttl":        record.TTL,
		"recordtype": record.RecordType,
		// active no longer used
		"active": record.Active,
		"target": record.Target,
	}
}

func (p *dns) NewRecordBody(ctx context.Context, params RecordBody) *RecordBody {

	logger := p.Log(ctx)
	logger.Debug("NewRecordBody")

	recordbody := &RecordBody{Name: params.Name}
	return recordbody
}

// Eval option lock arg passed into writable endpoints. Default is true, e.g. lock
func localLock(lockArg []bool) bool {

	for _, lock := range lockArg {
		// should only be one entry
		return lock
	}

	return true

}

func (p *dns) CreateRecord(ctx context.Context, record *RecordBody, zone string, recLock ...bool) error {
	// This lock will restrict the concurrency of API calls
	// to 1 save request at a time. This is needed for the Soa.Serial value which
	// is required to be incremented for every subsequent update to a zone
	// so we have to save just one request at a time to ensure this is always
	// incremented properly

	if localLock(recLock) {
		zoneRecordWriteLock.Lock()
		defer zoneRecordWriteLock.Unlock()
	}

	logger := p.Log(ctx)
	logger.Debug("CreateRecord")

	if err := record.Validate; err != nil {
		logger.Errorf("Record content not vaiid: %w", err)
		return nil
	}

	reqbody, err := convertStructToReqBody(record)
	if err != nil {
		return fmt.Errorf("failed to generate request body: %w", err)
	}

	var rec RecordBody
	postURL := fmt.Sprintf("/config-dns/v2/zones/%s/names/%s/types/%s", zone, record.Name, record.RecordType)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, postURL, reqbody)
	if err != nil {
		return fmt.Errorf("failed to create CreateRecord request: %w", err)
	}

	resp, err := p.Exec(req, &rec)
	if err != nil {
		return fmt.Errorf("CreateRecord request failed: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return session.NewAPIError(resp, logger)
	}

	return nil
}

func (p *dns) UpdateRecord(ctx context.Context, record *RecordBody, zone string, recLock ...bool) error {
	// This lock will restrict the concurrency of API calls
	// to 1 save request at a time. This is needed for the Soa.Serial value which
	// is required to be incremented for every subsequent update to a zone
	// so we have to save just one request at a time to ensure this is always
	// incremented properly

	if localLock(recLock) {
		zoneRecordWriteLock.Lock()
		defer zoneRecordWriteLock.Unlock()
	}

	logger := p.Log(ctx)
	logger.Debug("UpdateRecord")

	if err := record.Validate; err != nil {
		logger.Errorf("Record content not vaiid: %w", err)
		return nil
	}

	reqbody, err := convertStructToReqBody(record)
	if err != nil {
		return fmt.Errorf("failed to generate request body: %w", err)
	}

	var rec RecordBody
	putURL := fmt.Sprintf("/config-dns/v2/zones/%s/names/%s/types/%s", zone, record.Name, record.RecordType)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, putURL, reqbody)
	if err != nil {
		return fmt.Errorf("failed to create UpdateRecord request: %w", err)
	}

	resp, err := p.Exec(req, &rec)
	if err != nil {
		return fmt.Errorf("UpdateRecord request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return session.NewAPIError(resp, logger)
	}

	return nil
}

func (p *dns) DeleteRecord(ctx context.Context, record *RecordBody, zone string, recLock ...bool) error {
	// This lock will restrict the concurrency of API calls
	// to 1 save request at a time. This is needed for the Soa.Serial value which
	// is required to be incremented for every subsequent update to a zone
	// so we have to save just one request at a time to ensure this is always
	// incremented properly

	if localLock(recLock) {
		zoneRecordWriteLock.Lock()
		defer zoneRecordWriteLock.Unlock()
	}

	logger := p.Log(ctx)
	logger.Debug("DeleteRecord")

	if err := record.Validate; err != nil {
		logger.Errorf("Record content not vaiid: %w", err)
		return nil
	}

	var mtbody string
	deleteURL := fmt.Sprintf("/config-dns/v2/zones/%s/names/%s/types/%s", zone, record.Name, record.RecordType)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create DeleteRecord request: %w", err)
	}

	resp, err := p.Exec(req, &mtbody)
	if err != nil {
		return fmt.Errorf("DeleteRecord request failed: %w", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		return session.NewAPIError(resp, logger)
	}

	return nil
}
