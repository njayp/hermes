package client

type AddCNAMERequest struct {
	ZoneID  string `json:"zone_id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type DelCNAMERequest struct {
	ZoneID   string `json:"zone_id"`
	RecordID string `json:"record_id"`
}

type GetZoneIDRequest struct {
	Name string `json:"zone_name"`
}
