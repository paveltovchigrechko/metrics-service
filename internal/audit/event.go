package audit

// Event describes a successful metrics update. Receivers get it encoded as
// JSON:
//
//	{"ts":1760620000,"metrics":["Alloc","PollCount"],"ip_address":"192.168.0.42"}
type Event struct {
	// TS is the time of the update as a Unix timestamp in seconds.
	TS int64 `json:"ts"`
	// Metrics lists the names of the updated metrics, in request order.
	Metrics []string `json:"metrics"`
	// IP is the client address without the port.
	IP string `json:"ip_address"`
}
