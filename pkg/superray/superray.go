package superray

/*
#include <stdlib.h>
#include "superray.h"
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"unsafe"
)

// Response represents the standard JSON response from SuperRay
type Response struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// Server represents a proxy server
type Server struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	Port     int    `json:"port"`
	UUID     string `json:"uuid,omitempty"`
	Password string `json:"password,omitempty"`
	Method   string `json:"method,omitempty"`
	Network  string `json:"network,omitempty"`
	TLS      string `json:"tls,omitempty"`
	SNI      string `json:"sni,omitempty"`
	Path     string `json:"path,omitempty"`
	Host     string `json:"host,omitempty"`
	Flow     string `json:"flow,omitempty"`
	Security string `json:"security,omitempty"`
	// Reality fields
	PublicKey   string `json:"publicKey,omitempty"`
	ShortID     string `json:"shortId,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	// Additional fields
	Latency    int64  `json:"latency_ms"`           // -1 = not tested, 0 = timeout
	AlterId    int    `json:"alter_id,omitempty"`   // VMess alter ID
	ALPN       string `json:"alpn,omitempty"`       // ALPN protocols
	HeaderType string `json:"header_type,omitempty"` // TCP header type
	// Raw share link
	Link string `json:"link,omitempty"`
}

// InOutStats represents stats for an inbound or outbound
type InOutStats struct {
	Uplink   int64 `json:"uplink"`
	Downlink int64 `json:"downlink"`
}

// TrafficStats represents traffic statistics
type TrafficStats struct {
	Uplink       int64   `json:"uplink"`
	Downlink     int64   `json:"downlink"`
	UplinkRate   float64 `json:"uplink_rate"`
	DownlinkRate float64 `json:"downlink_rate"`
	// For backwards compatibility
	Upload       int64   `json:"upload"`
	Download     int64   `json:"download"`
	UploadRate   float64 `json:"-"`
	DownloadRate float64 `json:"-"`
	// Detailed per-tag stats
	Inbounds  map[string]InOutStats `json:"inbounds,omitempty"`
	Outbounds map[string]InOutStats `json:"outbounds,omitempty"`
}

// InstanceInfo represents Xray instance information
type InstanceInfo struct {
	ID     string `json:"id"`
	State  string `json:"state"`
	Uptime int64  `json:"uptime_seconds,omitempty"`
}

// LatencyResult represents a latency test result
type LatencyResult struct {
	Address   string  `json:"address"`
	Port      int     `json:"port"`
	Name      string  `json:"name,omitempty"`
	Latency   int     `json:"latency_ms"`
	AvgLatency float64 `json:"avg_latency_ms,omitempty"`
	MinLatency int     `json:"min_latency_ms,omitempty"`
	MaxLatency int     `json:"max_latency_ms,omitempty"`
	Success   bool    `json:"success"`
	Error     string  `json:"error,omitempty"`
}

// freeAndGetString frees C string and returns Go string
func freeAndGetString(cstr *C.char) string {
	if cstr == nil {
		return ""
	}
	goStr := C.GoString(cstr)
	C.SuperRay_Free(cstr)
	return goStr
}

// parseResponse parses the JSON response
func parseResponse(jsonStr string) (*Response, error) {
	var resp Response
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &resp, nil
}

// Version returns the SuperRay library version
func Version() (string, error) {
	result := freeAndGetString(C.SuperRay_Version())
	resp, err := parseResponse(result)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}
	var data struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return "", err
	}
	return data.Version, nil
}

// XrayVersion returns the underlying Xray-core version
func XrayVersion() (string, error) {
	result := freeAndGetString(C.SuperRay_XrayVersion())
	resp, err := parseResponse(result)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}
	var data struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return "", err
	}
	return data.Version, nil
}

// Run creates and starts an Xray instance
func Run(configJSON string) (string, error) {
	cConfig := C.CString(configJSON)
	defer C.free(unsafe.Pointer(cConfig))

	result := freeAndGetString(C.SuperRay_Run(cConfig))
	resp, err := parseResponse(result)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}
	var data struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return "", err
	}
	return data.ID, nil
}

// StopInstance stops a running Xray instance
func StopInstance(instanceID string) error {
	cID := C.CString(instanceID)
	defer C.free(unsafe.Pointer(cID))

	result := freeAndGetString(C.SuperRay_StopInstance(cID))
	resp, err := parseResponse(result)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

// DestroyInstance stops and destroys an Xray instance
func DestroyInstance(instanceID string) error {
	cID := C.CString(instanceID)
	defer C.free(unsafe.Pointer(cID))

	result := freeAndGetString(C.SuperRay_DestroyInstance(cID))
	resp, err := parseResponse(result)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

// StopAll stops all running instances
func StopAll() error {
	result := freeAndGetString(C.SuperRay_StopAll())
	resp, err := parseResponse(result)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

// GetInstanceState gets the state of an instance
func GetInstanceState(instanceID string) (string, error) {
	cID := C.CString(instanceID)
	defer C.free(unsafe.Pointer(cID))

	result := freeAndGetString(C.SuperRay_GetInstanceState(cID))
	resp, err := parseResponse(result)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}
	var data struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return "", err
	}
	return data.State, nil
}

// ListInstances returns all instance IDs
func ListInstances() ([]string, error) {
	result := freeAndGetString(C.SuperRay_ListInstances())
	resp, err := parseResponse(result)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}
	var data struct {
		Instances []string `json:"instances"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, err
	}
	return data.Instances, nil
}

// SetAssetDir sets the geo asset directory
func SetAssetDir(dir string) error {
	cDir := C.CString(dir)
	defer C.free(unsafe.Pointer(cDir))

	result := freeAndGetString(C.SuperRay_SetAssetDir(cDir))
	resp, err := parseResponse(result)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

// ParseShareLink parses a single share link
func ParseShareLink(link string) (*Server, error) {
	cLink := C.CString(link)
	defer C.free(unsafe.Pointer(cLink))

	result := freeAndGetString(C.SuperRay_ParseShareLink(cLink))
	resp, err := parseResponse(result)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}
	var server Server
	if err := json.Unmarshal(resp.Data, &server); err != nil {
		return nil, err
	}
	server.Link = link
	return &server, nil
}

// ParseShareLinks parses multiple share links
func ParseShareLinks(content string) ([]*Server, error) {
	cContent := C.CString(content)
	defer C.free(unsafe.Pointer(cContent))

	result := freeAndGetString(C.SuperRay_ParseShareLinks(cContent))
	resp, err := parseResponse(result)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}
	var data struct {
		Links  []json.RawMessage `json:"links"`
		Count  int               `json:"count"`
		Errors []string          `json:"errors"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, err
	}

	servers := make([]*Server, 0, len(data.Links))
	for _, linkData := range data.Links {
		var server Server
		if err := json.Unmarshal(linkData, &server); err != nil {
			continue
		}
		servers = append(servers, &server)
	}
	return servers, nil
}

// ShareLinkToXrayConfig converts a share link to Xray outbound config
func ShareLinkToXrayConfig(link string) (string, error) {
	cLink := C.CString(link)
	defer C.free(unsafe.Pointer(cLink))

	result := freeAndGetString(C.SuperRay_ShareLinkToXrayConfig(cLink))
	resp, err := parseResponse(result)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}
	return string(resp.Data), nil
}

// CreateClientConfig creates a complete client configuration
func CreateClientConfig(localPort int, outboundJSON string) (string, error) {
	cOutbound := C.CString(outboundJSON)
	defer C.free(unsafe.Pointer(cOutbound))

	result := freeAndGetString(C.SuperRay_CreateClientConfig(C.int(localPort), cOutbound))
	resp, err := parseResponse(result)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}
	return string(resp.Data), nil
}

// TCPPing tests TCP connectivity
func TCPPing(address string, port int, timeoutMs int) (int, error) {
	cAddr := C.CString(address)
	defer C.free(unsafe.Pointer(cAddr))

	result := freeAndGetString(C.SuperRay_TCPPing(cAddr, C.int(port), C.int(timeoutMs)))
	resp, err := parseResponse(result)
	if err != nil {
		return 0, err
	}
	if !resp.Success {
		return 0, fmt.Errorf(resp.Error)
	}
	var data struct {
		LatencyMs int `json:"latency_ms"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return 0, err
	}
	return data.LatencyMs, nil
}

// TCPPingMultiple tests TCP connectivity multiple times
func TCPPingMultiple(address string, port int, count int, timeoutMs int) (*LatencyResult, error) {
	cAddr := C.CString(address)
	defer C.free(unsafe.Pointer(cAddr))

	result := freeAndGetString(C.SuperRay_TCPPingMultiple(cAddr, C.int(port), C.int(count), C.int(timeoutMs)))
	resp, err := parseResponse(result)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}
	var data LatencyResult
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, err
	}
	data.Success = true
	return &data, nil
}

// BatchLatencyTest tests latency for multiple servers
func BatchLatencyTest(servers []map[string]interface{}, concurrent int, count int, timeoutMs int) ([]LatencyResult, error) {
	serversJSON, err := json.Marshal(servers)
	if err != nil {
		return nil, err
	}

	cServers := C.CString(string(serversJSON))
	defer C.free(unsafe.Pointer(cServers))

	result := freeAndGetString(C.SuperRay_BatchLatencyTest(cServers, C.int(concurrent), C.int(count), C.int(timeoutMs)))
	resp, err := parseResponse(result)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}
	var data struct {
		Results []LatencyResult `json:"results"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, err
	}
	return data.Results, nil
}

// GetXrayStats gets traffic statistics from Xray core
func GetXrayStats() (*TrafficStats, error) {
	result := freeAndGetString(C.SuperRay_GetXrayStats())
	resp, err := parseResponse(result)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}
	var stats TrafficStats
	if err := json.Unmarshal(resp.Data, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetCurrentSpeed gets current upload/download speed
func GetCurrentSpeed() (*TrafficStats, error) {
	result := freeAndGetString(C.SuperRay_GetCurrentSpeed())
	resp, err := parseResponse(result)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}
	var stats TrafficStats
	if err := json.Unmarshal(resp.Data, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// AddSubscription adds a new subscription
func AddSubscription(name, url string) error {
	cName := C.CString(name)
	cURL := C.CString(url)
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cURL))

	result := freeAndGetString(C.SuperRay_AddSubscription(cName, cURL))
	resp, err := parseResponse(result)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

// UpdateSubscription updates a subscription
func UpdateSubscription(name string) ([]*Server, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	result := freeAndGetString(C.SuperRay_UpdateSubscription(cName))
	resp, err := parseResponse(result)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}
	var data struct {
		Servers []*Server `json:"servers"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, err
	}
	return data.Servers, nil
}

// GetAllServers gets all servers from all subscriptions
func GetAllServers() ([]*Server, error) {
	result := freeAndGetString(C.SuperRay_GetAllServers())
	resp, err := parseResponse(result)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}
	var data struct {
		Servers []*Server `json:"servers"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, err
	}
	return data.Servers, nil
}

// GetRecentLogs gets recent log entries
func GetRecentLogs(count int) ([]string, error) {
	result := freeAndGetString(C.SuperRay_GetRecentLogs(C.int(count)))
	resp, err := parseResponse(result)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}
	var data struct {
		Logs []string `json:"logs"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, err
	}
	return data.Logs, nil
}

// SetLogLevel sets the log level
func SetLogLevel(level string) error {
	cLevel := C.CString(level)
	defer C.free(unsafe.Pointer(cLevel))

	result := freeAndGetString(C.SuperRay_SetLogLevel(cLevel))
	resp, err := parseResponse(result)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

// HTTPPing tests HTTP connectivity through proxy
func HTTPPing(url, proxyAddr string, timeoutMs int) (int, error) {
	cURL := C.CString(url)
	cProxy := C.CString(proxyAddr)
	defer C.free(unsafe.Pointer(cURL))
	defer C.free(unsafe.Pointer(cProxy))

	result := freeAndGetString(C.SuperRay_HTTPPing(cURL, cProxy, C.int(timeoutMs)))
	resp, err := parseResponse(result)
	if err != nil {
		return 0, err
	}
	if !resp.Success {
		return 0, fmt.Errorf(resp.Error)
	}
	var data struct {
		LatencyMs int `json:"latency_ms"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return 0, err
	}
	return data.LatencyMs, nil
}

// CreateSOCKSInbound creates a SOCKS5 inbound configuration
func CreateSOCKSInbound(tag, listen string, port int) (string, error) {
	cTag := C.CString(tag)
	cListen := C.CString(listen)
	defer C.free(unsafe.Pointer(cTag))
	defer C.free(unsafe.Pointer(cListen))

	result := freeAndGetString(C.SuperRay_CreateSOCKSInbound(cTag, cListen, C.int(port)))
	resp, err := parseResponse(result)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}
	return string(resp.Data), nil
}

// CreateHTTPInbound creates an HTTP proxy inbound configuration
func CreateHTTPInbound(tag, listen string, port int) (string, error) {
	cTag := C.CString(tag)
	cListen := C.CString(listen)
	defer C.free(unsafe.Pointer(cTag))
	defer C.free(unsafe.Pointer(cListen))

	result := freeAndGetString(C.SuperRay_CreateHTTPInbound(cTag, cListen, C.int(port)))
	resp, err := parseResponse(result)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}
	return string(resp.Data), nil
}

// CreateFreedomOutbound creates a direct outbound configuration
func CreateFreedomOutbound(tag string) (string, error) {
	cTag := C.CString(tag)
	defer C.free(unsafe.Pointer(cTag))

	result := freeAndGetString(C.SuperRay_CreateFreedomOutbound(cTag))
	resp, err := parseResponse(result)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}
	return string(resp.Data), nil
}

// BuildFullConfig builds a complete Xray configuration
func BuildFullConfig(inboundsJSON, outboundsJSON, logLevel, dnsServersJSON string) (string, error) {
	cInbounds := C.CString(inboundsJSON)
	cOutbounds := C.CString(outboundsJSON)
	cLogLevel := C.CString(logLevel)
	cDNS := C.CString(dnsServersJSON)
	defer C.free(unsafe.Pointer(cInbounds))
	defer C.free(unsafe.Pointer(cOutbounds))
	defer C.free(unsafe.Pointer(cLogLevel))
	defer C.free(unsafe.Pointer(cDNS))

	result := freeAndGetString(C.SuperRay_BuildFullConfig(cInbounds, cOutbounds, cLogLevel, cDNS))
	resp, err := parseResponse(result)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}
	return string(resp.Data), nil
}

// GetFreePorts returns available TCP ports
func GetFreePorts(count int) ([]int, error) {
	result := freeAndGetString(C.SuperRay_GetFreePorts(C.int(count)))
	resp, err := parseResponse(result)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}
	var data struct {
		Ports []int `json:"ports"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, err
	}
	return data.Ports, nil
}

// CreateRoutingRuleDomain creates a domain-based routing rule
func CreateRoutingRuleDomain(domains []string, outboundTag string) (string, error) {
	domainsJSON, _ := json.Marshal(domains)
	cDomains := C.CString(string(domainsJSON))
	cTag := C.CString(outboundTag)
	defer C.free(unsafe.Pointer(cDomains))
	defer C.free(unsafe.Pointer(cTag))

	result := freeAndGetString(C.SuperRay_CreateRoutingRuleDomain(cDomains, cTag))
	resp, err := parseResponse(result)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}
	return string(resp.Data), nil
}

// CreateRoutingRuleIP creates an IP-based routing rule
func CreateRoutingRuleIP(ips []string, outboundTag string) (string, error) {
	ipsJSON, _ := json.Marshal(ips)
	cIPs := C.CString(string(ipsJSON))
	cTag := C.CString(outboundTag)
	defer C.free(unsafe.Pointer(cIPs))
	defer C.free(unsafe.Pointer(cTag))

	result := freeAndGetString(C.SuperRay_CreateRoutingRuleIP(cIPs, cTag))
	resp, err := parseResponse(result)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}
	return string(resp.Data), nil
}

// ========== TUN Device Functions ==========

// CreateTUNInbound creates a TUN inbound configuration
func CreateTUNInbound(tag string, addresses []string, mtu int) (string, error) {
	addressesJSON, _ := json.Marshal(addresses)
	cTag := C.CString(tag)
	cAddresses := C.CString(string(addressesJSON))
	defer C.free(unsafe.Pointer(cTag))
	defer C.free(unsafe.Pointer(cAddresses))

	result := freeAndGetString(C.SuperRay_CreateTUNInbound(cTag, cAddresses, C.int(mtu)))
	resp, err := parseResponse(result)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}
	return string(resp.Data), nil
}

// CreateTUNInboundFull creates a TUN inbound with full options
func CreateTUNInboundFull(tag, name string, addresses []string, mtu int, autoRoute bool) (string, error) {
	addressesJSON, _ := json.Marshal(addresses)
	cTag := C.CString(tag)
	cName := C.CString(name)
	cAddresses := C.CString(string(addressesJSON))
	defer C.free(unsafe.Pointer(cTag))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cAddresses))

	autoRouteInt := 0
	if autoRoute {
		autoRouteInt = 1
	}

	result := freeAndGetString(C.SuperRay_CreateTUNInboundFull(cTag, cName, cAddresses, C.int(mtu), C.int(autoRouteInt)))
	resp, err := parseResponse(result)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}
	return string(resp.Data), nil
}

// CreateTUNDevice creates a TUN device
func CreateTUNDevice(config map[string]interface{}) (string, error) {
	configJSON, _ := json.Marshal(config)
	cConfig := C.CString(string(configJSON))
	defer C.free(unsafe.Pointer(cConfig))

	result := freeAndGetString(C.SuperRay_CreateTUNDevice(cConfig))
	resp, err := parseResponse(result)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}
	return string(resp.Data), nil
}

// RemoveTUNDevice removes a TUN device
func RemoveTUNDevice(tag string) error {
	cTag := C.CString(tag)
	defer C.free(unsafe.Pointer(cTag))

	result := freeAndGetString(C.SuperRay_RemoveTUNDevice(cTag))
	resp, err := parseResponse(result)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

// CloseAllTUNDevices closes all TUN devices
func CloseAllTUNDevices() error {
	result := freeAndGetString(C.SuperRay_CloseAllTUNDevices())
	resp, err := parseResponse(result)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

// CreateCallbackTUNWithDialer creates a TUN device connected to Xray instance
func CreateCallbackTUNWithDialer(config map[string]interface{}, instanceID, outboundTag string) (string, error) {
	configJSON, _ := json.Marshal(config)
	cConfig := C.CString(string(configJSON))
	cInstanceID := C.CString(instanceID)
	cOutboundTag := C.CString(outboundTag)
	defer C.free(unsafe.Pointer(cConfig))
	defer C.free(unsafe.Pointer(cInstanceID))
	defer C.free(unsafe.Pointer(cOutboundTag))

	result := freeAndGetString(C.SuperRay_CreateCallbackTUNWithDialer(cConfig, cInstanceID, cOutboundTag))
	resp, err := parseResponse(result)
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf(resp.Error)
	}
	return string(resp.Data), nil
}

// StartCallbackTUN starts a callback TUN device
func StartCallbackTUN(tag string) error {
	cTag := C.CString(tag)
	defer C.free(unsafe.Pointer(cTag))

	result := freeAndGetString(C.SuperRay_StartCallbackTUN(cTag))
	resp, err := parseResponse(result)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

// StopCallbackTUN stops and removes a callback TUN device
func StopCallbackTUN(tag string) error {
	cTag := C.CString(tag)
	defer C.free(unsafe.Pointer(cTag))

	result := freeAndGetString(C.SuperRay_StopCallbackTUN(cTag))
	resp, err := parseResponse(result)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

// CloseAllCallbackTUNs closes all callback TUN devices
func CloseAllCallbackTUNs() error {
	result := freeAndGetString(C.SuperRay_CloseAllCallbackTUNs())
	resp, err := parseResponse(result)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

// ========== GeoIP Functions ==========

// GeoIPInfo represents GeoIP lookup result
type GeoIPInfo struct {
	IP          string `json:"ip"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Region      string `json:"region"`
	City        string `json:"city"`
	ASN         int    `json:"asn"`
	ASOrg       string `json:"as_org"`
}

// LookupHost resolves hostname to IP addresses
func LookupHost(host string) ([]string, error) {
	cHost := C.CString(host)
	defer C.free(unsafe.Pointer(cHost))

	result := freeAndGetString(C.SuperRay_LookupHost(cHost))
	resp, err := parseResponse(result)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}
	var data struct {
		Addresses []string `json:"addresses"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, err
	}
	return data.Addresses, nil
}

// ========== System TUN Functions (Desktop) ==========

// SystemTUNInfo represents system TUN device information
type SystemTUNInfo struct {
	Tag    string `json:"tag"`
	Name   string `json:"name"`
	MTU    int    `json:"mtu"`
	Status string `json:"status"`
}

// CreateSystemTUN creates a system-level TUN device (requires root/admin)
func CreateSystemTUN(tag string, addresses []string, mtu int) (*SystemTUNInfo, error) {
	config := map[string]interface{}{
		"tag":       tag,
		"mtu":       mtu,
		"addresses": addresses,
	}
	configJSON, _ := json.Marshal(config)
	cConfig := C.CString(string(configJSON))
	defer C.free(unsafe.Pointer(cConfig))

	result := freeAndGetString(C.SuperRay_CreateSystemTUN(cConfig))
	resp, err := parseResponse(result)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	var info SystemTUNInfo
	if err := json.Unmarshal(resp.Data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// StartSystemTUNStack starts the TUN stack connected to Xray instance
func StartSystemTUNStack(tag, instanceID, outboundTag string) error {
	cTag := C.CString(tag)
	defer C.free(unsafe.Pointer(cTag))
	cInstanceID := C.CString(instanceID)
	defer C.free(unsafe.Pointer(cInstanceID))
	cOutboundTag := C.CString(outboundTag)
	defer C.free(unsafe.Pointer(cOutboundTag))

	result := freeAndGetString(C.SuperRay_StartSystemTUNStack(cTag, cInstanceID, cOutboundTag))
	resp, err := parseResponse(result)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

// SetupRoutes sets up system routes for TUN
func SetupRoutes(tag, serverAddress string) error {
	cTag := C.CString(tag)
	defer C.free(unsafe.Pointer(cTag))
	cServerAddr := C.CString(serverAddress)
	defer C.free(unsafe.Pointer(cServerAddr))

	result := freeAndGetString(C.SuperRay_SetupRoutes(cTag, cServerAddr))
	resp, err := parseResponse(result)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

// CleanupRoutes cleans up system routes
func CleanupRoutes(tag string) error {
	cTag := C.CString(tag)
	defer C.free(unsafe.Pointer(cTag))

	result := freeAndGetString(C.SuperRay_CleanupRoutes(cTag))
	resp, err := parseResponse(result)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

// CloseSystemTUN closes a system TUN device
func CloseSystemTUN(tag string) error {
	cTag := C.CString(tag)
	defer C.free(unsafe.Pointer(cTag))

	result := freeAndGetString(C.SuperRay_CloseSystemTUN(cTag))
	resp, err := parseResponse(result)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

// CloseAllSystemTUNs closes all system TUN devices
func CloseAllSystemTUNs() error {
	result := freeAndGetString(C.SuperRay_CloseAllSystemTUNs())
	resp, err := parseResponse(result)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}
	return nil
}
