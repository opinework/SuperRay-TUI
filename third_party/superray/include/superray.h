/*
 * SuperRay - Cross-platform Xray Library with C ABI
 *
 * This header provides the C API for SuperRay library.
 * All functions that return char* require caller to free the memory using SuperRay_Free().
 * All returned strings are JSON formatted.
 *
 * Response format:
 * {
 *   "success": true|false,
 *   "data": {...},      // present on success
 *   "error": "..."      // present on failure
 * }
 */

#ifndef SUPERRAY_H
#define SUPERRAY_H

#include <stdbool.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/* ========== Version Functions ========== */

/* Get SuperRay library version */
extern char* SuperRay_Version(void);

/* Get underlying Xray-core version */
extern char* SuperRay_XrayVersion(void);

/* ========== Instance Management ========== */

/*
 * Create a new Xray instance from JSON config
 * @param configJSON: Full Xray JSON configuration string
 * @return JSON: {"success":true,"data":{"id":"instance_id"}}
 */
extern char* SuperRay_CreateInstance(const char* configJSON);

/*
 * Start an Xray instance by ID
 * @param instanceID: Instance ID returned from CreateInstance
 * @return JSON with success status
 */
extern char* SuperRay_StartInstance(const char* instanceID);

/*
 * Stop a running Xray instance
 * @param instanceID: Instance ID
 * @return JSON with success status
 */
extern char* SuperRay_StopInstance(const char* instanceID);

/*
 * Stop and destroy an Xray instance
 * @param instanceID: Instance ID
 * @return JSON with success status
 */
extern char* SuperRay_DestroyInstance(const char* instanceID);

/*
 * Get the state of an instance
 * @param instanceID: Instance ID
 * @return JSON: {"success":true,"data":{"id":"...","state":"running|stopped|starting|stopping"}}
 */
extern char* SuperRay_GetInstanceState(const char* instanceID);

/*
 * Get detailed information about an instance
 * @param instanceID: Instance ID
 * @return JSON with id, state, start_at, uptime_seconds
 */
extern char* SuperRay_GetInstanceInfo(const char* instanceID);

/*
 * List all instance IDs
 * @return JSON: {"success":true,"data":{"instances":["id1","id2"],"count":2}}
 */
extern char* SuperRay_ListInstances(void);

/* ========== Simple API ========== */

/*
 * Create, start and run Xray in one call
 * @param configJSON: Full Xray JSON configuration
 * @return JSON with instance ID and status
 */
extern char* SuperRay_Run(const char* configJSON);

/*
 * Run Xray from a config file path (supports JSON, YAML, TOML)
 * @param configPath: Path to Xray config file
 * @return JSON with instance ID and status
 */
extern char* SuperRay_RunFromFile(const char* configPath);

/*
 * Run Xray from multiple config files (like: xray run -c a.json -c b.json)
 * @param pathsJSON: JSON array of config file paths, e.g. ["/path/a.json", "/path/b.json"]
 * @return JSON with instance ID and status
 */
extern char* SuperRay_RunFromFiles(const char* pathsJSON);

/*
 * Run Xray from all config files in a directory
 * @param configDir: Directory path containing config files
 * @return JSON with instance ID and status
 */
extern char* SuperRay_RunFromDir(const char* configDir);

/*
 * Stop all running instances
 * @return JSON with count of stopped instances
 */
extern char* SuperRay_StopAll(void);

/*
 * Validate Xray configuration without starting
 * @param configJSON: Xray JSON configuration
 * @return JSON: {"success":true,"data":{"valid":true}}
 */
extern char* SuperRay_ValidateConfig(const char* configJSON);

/* ========== DNS Functions ========== */

/*
 * Initialize custom DNS servers
 * @param serversJSON: JSON array of DNS servers, e.g. ["8.8.8.8","1.1.1.1"]
 * @return JSON with success status
 */
extern char* SuperRay_InitDNS(const char* serversJSON);

/*
 * Reset to system default DNS
 * @return JSON with success status
 */
extern char* SuperRay_ResetDNS(void);

/*
 * Resolve hostname to IP addresses
 * @param host: Hostname to resolve
 * @return JSON: {"success":true,"data":{"host":"...","addresses":["1.2.3.4"]}}
 */
extern char* SuperRay_LookupHost(const char* host);

/* ========== Share Link Functions ========== */

/*
 * Parse a single share link (vmess://, vless://, trojan://, ss://)
 * @param link: Share link string
 * @return JSON with parsed link details
 */
extern char* SuperRay_ParseShareLink(const char* link);

/*
 * Parse multiple share links (one per line)
 * @param content: Multi-line string with share links
 * @return JSON: {"success":true,"data":{"links":[...],"errors":[...],"count":N}}
 */
extern char* SuperRay_ParseShareLinks(const char* content);

/*
 * Convert a share link to Xray outbound config
 * @param link: Share link string
 * @return JSON with Xray outbound configuration
 */
extern char* SuperRay_ShareLinkToXrayConfig(const char* link);

/*
 * Generate a share link from config
 * @param protocol: Protocol name (vmess, vless, trojan, ss)
 * @param configJSON: JSON object with address, port, uuid, etc.
 * @return JSON: {"success":true,"data":{"link":"vmess://..."}}
 */
extern char* SuperRay_GenerateShareLink(const char* protocol, const char* configJSON);

/*
 * Convert multiple share links to Xray config with outbounds
 * @param content: Multi-line share links
 * @return JSON Xray config with outbounds array
 */
extern char* SuperRay_ConvertLinksToConfig(const char* content);

/* ========== Geo Data Functions ========== */

/*
 * Set the asset directory for geo files (geoip.dat, geosite.dat)
 * @param dir: Directory path
 * @return JSON with success status
 */
extern char* SuperRay_SetAssetDir(const char* dir);

/*
 * Get the current asset directory
 * @return JSON: {"success":true,"data":{"asset_dir":"..."}}
 */
extern char* SuperRay_GetAssetDir(void);

/*
 * Check if geo files exist
 * @return JSON with geoip_path and geosite_path
 */
extern char* SuperRay_CheckGeoFiles(void);

/*
 * Find geo references in a config
 * @param configJSON: Xray JSON configuration
 * @return JSON: {"success":true,"data":{"references":["geoip:cn"],"count":1}}
 */
extern char* SuperRay_FindGeoInConfig(const char* configJSON);

/* ========== Network Utility Functions ========== */

/*
 * Get available TCP ports
 * @param count: Number of ports to find (max 100)
 * @return JSON: {"success":true,"data":{"ports":[12345,12346],"count":2}}
 */
extern char* SuperRay_GetFreePorts(int count);

/*
 * TCP ping to test connectivity
 * @param address: Address in format "host:port"
 * @param timeoutMs: Timeout in milliseconds (0 = default 5000ms)
 * @return JSON with latency_ms
 */
extern char* SuperRay_Ping(const char* address, int timeoutMs);

/*
 * HTTP ping through optional proxy
 * @param url: URL to ping (e.g., "https://www.google.com")
 * @param proxyAddr: Proxy address "host:port" or empty for direct
 * @param timeoutMs: Timeout in milliseconds (0 = default 10000ms)
 * @return JSON with status_code and latency_ms
 */
extern char* SuperRay_HTTPPing(const char* url, const char* proxyAddr, int timeoutMs);

/*
 * Check if a port is open
 * @param host: Host address
 * @param port: Port number
 * @param timeoutMs: Timeout in milliseconds (0 = default 3000ms)
 * @return JSON: {"success":true,"data":{"host":"...","port":443,"open":true}}
 */
extern char* SuperRay_CheckPort(const char* host, int port, int timeoutMs);

/* ========== Config Builder Functions ========== */

/*
 * Create a quick proxy configuration
 * @param localPort: Local SOCKS5 port
 * @param protocol: Protocol (vmess, vless, trojan, ss)
 * @param address: Server address
 * @param port: Server port
 * @param uuid: UUID or password
 * @return JSON with generated config
 */
extern char* SuperRay_QuickConfig(int localPort, const char* protocol, const char* address, int port, const char* uuid);

/*
 * Build a detailed config from parameters
 * @param paramsJSON: JSON with local_port, protocol, address, port, uuid, password, method, network, tls, sni, path, host
 * @return JSON with generated config
 */
extern char* SuperRay_BuildConfig(const char* paramsJSON);

/*
 * Merge outbounds into a base config
 * @param baseConfigJSON: Base Xray config JSON
 * @param outboundsJSON: Array of outbound configs to add
 * @return JSON with merged config
 */
extern char* SuperRay_MergeConfigs(const char* baseConfigJSON, const char* outboundsJSON);

/* ========== Protocol Inbound Builders ========== */

/*
 * Create a SOCKS5 inbound configuration
 * @param tag: Inbound tag name
 * @param listen: Listen address (e.g., "127.0.0.1")
 * @param port: Listen port
 * @return JSON with inbound configuration
 */
extern char* SuperRay_CreateSOCKSInbound(const char* tag, const char* listen, int port);

/*
 * Create a SOCKS5 inbound with authentication
 * @param tag: Inbound tag name
 * @param listen: Listen address
 * @param port: Listen port
 * @param user: Username
 * @param pass: Password
 * @return JSON with inbound configuration
 */
extern char* SuperRay_CreateSOCKSInboundWithAuth(const char* tag, const char* listen, int port, const char* user, const char* pass);

/*
 * Create an HTTP proxy inbound configuration
 * @param tag: Inbound tag name
 * @param listen: Listen address
 * @param port: Listen port
 * @return JSON with inbound configuration
 */
extern char* SuperRay_CreateHTTPInbound(const char* tag, const char* listen, int port);

/*
 * Create an HTTP proxy inbound with authentication
 * @param tag: Inbound tag name
 * @param listen: Listen address
 * @param port: Listen port
 * @param user: Username
 * @param pass: Password
 * @return JSON with inbound configuration
 */
extern char* SuperRay_CreateHTTPInboundWithAuth(const char* tag, const char* listen, int port, const char* user, const char* pass);

/*
 * Create a dokodemo-door inbound (transparent proxy)
 * @param tag: Inbound tag name
 * @param listen: Listen address
 * @param port: Listen port
 * @param network: Network type ("tcp", "udp", "tcp,udp")
 * @return JSON with inbound configuration
 */
extern char* SuperRay_CreateDokodemoInbound(const char* tag, const char* listen, int port, const char* network);

/*
 * Create a dokodemo-door inbound forwarding to specific address
 * @param tag: Inbound tag name
 * @param listen: Listen address
 * @param port: Listen port
 * @param destAddr: Destination address
 * @param destPort: Destination port
 * @param network: Network type
 * @return JSON with inbound configuration
 */
extern char* SuperRay_CreateDokodemoInboundToAddr(const char* tag, const char* listen, int port, const char* destAddr, int destPort, const char* network);

/*
 * Create a VMess inbound configuration
 * @param tag: Inbound tag name
 * @param listen: Listen address
 * @param port: Listen port
 * @param uuid: User UUID
 * @return JSON with inbound configuration
 */
extern char* SuperRay_CreateVMessInbound(const char* tag, const char* listen, int port, const char* uuid);

/*
 * Create a VLESS inbound configuration
 * @param tag: Inbound tag name
 * @param listen: Listen address
 * @param port: Listen port
 * @param uuid: User UUID
 * @return JSON with inbound configuration
 */
extern char* SuperRay_CreateVLESSInbound(const char* tag, const char* listen, int port, const char* uuid);

/*
 * Create a VLESS inbound with XTLS
 * @param tag: Inbound tag name
 * @param listen: Listen address
 * @param port: Listen port
 * @param uuid: User UUID
 * @param flow: XTLS flow (e.g., "xtls-rprx-vision")
 * @return JSON with inbound configuration
 */
extern char* SuperRay_CreateVLESSInboundXTLS(const char* tag, const char* listen, int port, const char* uuid, const char* flow);

/*
 * Create a Trojan inbound configuration
 * @param tag: Inbound tag name
 * @param listen: Listen address
 * @param port: Listen port
 * @param password: User password
 * @return JSON with inbound configuration
 */
extern char* SuperRay_CreateTrojanInbound(const char* tag, const char* listen, int port, const char* password);

/*
 * Create a Shadowsocks inbound configuration
 * @param tag: Inbound tag name
 * @param listen: Listen address
 * @param port: Listen port
 * @param method: Encryption method (e.g., "aes-256-gcm")
 * @param password: Password
 * @return JSON with inbound configuration
 */
extern char* SuperRay_CreateShadowsocksInbound(const char* tag, const char* listen, int port, const char* method, const char* password);

/* ========== Protocol Outbound Builders ========== */

/*
 * Create a freedom (direct) outbound
 * @param tag: Outbound tag name
 * @return JSON with outbound configuration
 */
extern char* SuperRay_CreateFreedomOutbound(const char* tag);

/*
 * Create a blackhole outbound
 * @param tag: Outbound tag name
 * @return JSON with outbound configuration
 */
extern char* SuperRay_CreateBlackholeOutbound(const char* tag);

/*
 * Create a VMess outbound configuration
 * @param tag: Outbound tag name
 * @param address: Server address
 * @param port: Server port
 * @param uuid: User UUID
 * @param security: Security type ("auto", "aes-128-gcm", "chacha20-poly1305", "none")
 * @return JSON with outbound configuration
 */
extern char* SuperRay_CreateVMessOutbound(const char* tag, const char* address, int port, const char* uuid, const char* security);

/*
 * Create a VMess outbound with full options
 * @param tag: Outbound tag name
 * @param address: Server address
 * @param port: Server port
 * @param uuid: User UUID
 * @param security: Security type
 * @param network: Network type ("tcp", "ws", "grpc", "h2")
 * @param tls: Enable TLS (1=true, 0=false)
 * @param sni: Server Name Indication
 * @param path: WebSocket/gRPC path
 * @param host: Host header
 * @return JSON with outbound configuration
 */
extern char* SuperRay_CreateVMessOutboundFull(const char* tag, const char* address, int port, const char* uuid, const char* security, const char* network, int tls, const char* sni, const char* path, const char* host);

/*
 * Create a VLESS outbound configuration
 * @param tag: Outbound tag name
 * @param address: Server address
 * @param port: Server port
 * @param uuid: User UUID
 * @return JSON with outbound configuration
 */
extern char* SuperRay_CreateVLESSOutbound(const char* tag, const char* address, int port, const char* uuid);

/*
 * Create a VLESS outbound with XTLS
 * @param tag: Outbound tag name
 * @param address: Server address
 * @param port: Server port
 * @param uuid: User UUID
 * @param flow: XTLS flow (e.g., "xtls-rprx-vision")
 * @param sni: Server Name Indication
 * @param fingerprint: TLS fingerprint ("chrome", "firefox", "safari", "randomized")
 * @return JSON with outbound configuration
 */
extern char* SuperRay_CreateVLESSOutboundXTLS(const char* tag, const char* address, int port, const char* uuid, const char* flow, const char* sni, const char* fingerprint);

/*
 * Create a VLESS outbound with Reality
 * @param tag: Outbound tag name
 * @param address: Server address
 * @param port: Server port
 * @param uuid: User UUID
 * @param flow: XTLS flow
 * @param sni: Server Name Indication
 * @param fingerprint: TLS fingerprint
 * @param publicKey: Reality public key
 * @param shortId: Reality short ID
 * @return JSON with outbound configuration
 */
extern char* SuperRay_CreateVLESSOutboundReality(const char* tag, const char* address, int port, const char* uuid, const char* flow, const char* sni, const char* fingerprint, const char* publicKey, const char* shortId);

/*
 * Create a VLESS outbound with full options
 * @param tag: Outbound tag name
 * @param address: Server address
 * @param port: Server port
 * @param uuid: User UUID
 * @param flow: XTLS flow (empty for none)
 * @param network: Network type
 * @param security: Security type ("none", "tls", "reality")
 * @param sni: Server Name Indication
 * @param path: WebSocket/gRPC path
 * @param host: Host header
 * @return JSON with outbound configuration
 */
extern char* SuperRay_CreateVLESSOutboundFull(const char* tag, const char* address, int port, const char* uuid, const char* flow, const char* network, const char* security, const char* sni, const char* path, const char* host);

/*
 * Create a Trojan outbound configuration
 * @param tag: Outbound tag name
 * @param address: Server address
 * @param port: Server port
 * @param password: User password
 * @param sni: Server Name Indication (empty uses address)
 * @return JSON with outbound configuration
 */
extern char* SuperRay_CreateTrojanOutbound(const char* tag, const char* address, int port, const char* password, const char* sni);

/*
 * Create a Trojan outbound with full options
 * @param tag: Outbound tag name
 * @param address: Server address
 * @param port: Server port
 * @param password: User password
 * @param network: Network type ("tcp", "ws", "grpc")
 * @param sni: Server Name Indication
 * @param path: WebSocket/gRPC path
 * @return JSON with outbound configuration
 */
extern char* SuperRay_CreateTrojanOutboundFull(const char* tag, const char* address, int port, const char* password, const char* network, const char* sni, const char* path);

/*
 * Create a Shadowsocks outbound configuration
 * @param tag: Outbound tag name
 * @param address: Server address
 * @param port: Server port
 * @param method: Encryption method
 * @param password: Password
 * @return JSON with outbound configuration
 */
extern char* SuperRay_CreateShadowsocksOutbound(const char* tag, const char* address, int port, const char* method, const char* password);

/*
 * Create a WireGuard outbound configuration
 * @param tag: Outbound tag name
 * @param privateKey: WireGuard private key
 * @param addressJSON: JSON array of addresses, e.g. ["10.0.0.2/32", "fd00::2/128"]
 * @param peersJSON: JSON array of peer objects with publicKey, endpoint, allowedIPs
 * @return JSON with outbound configuration
 */
extern char* SuperRay_CreateWireGuardOutbound(const char* tag, const char* privateKey, const char* addressJSON, const char* peersJSON);

/*
 * Create a DNS outbound
 * @param tag: Outbound tag name
 * @return JSON with outbound configuration
 */
extern char* SuperRay_CreateDNSOutbound(const char* tag);

/* ========== Full Config Builder ========== */

/*
 * Create a complete client configuration
 * @param localPort: Local SOCKS5 port (HTTP port will be localPort+1)
 * @param outboundJSON: JSON object of outbound configuration
 * @return JSON with complete Xray configuration
 */
extern char* SuperRay_CreateClientConfig(int localPort, const char* outboundJSON);

/*
 * Build a full Xray configuration from components
 * @param inboundsJSON: JSON array of inbound configurations
 * @param outboundsJSON: JSON array of outbound configurations
 * @param logLevel: Log level ("debug", "info", "warning", "error", "none")
 * @param dnsServersJSON: JSON array of DNS servers (can be empty "[]")
 * @return JSON with complete Xray configuration
 */
extern char* SuperRay_BuildFullConfig(const char* inboundsJSON, const char* outboundsJSON, const char* logLevel, const char* dnsServersJSON);

/*
 * Get list of supported protocols
 * @return JSON with inbound, outbound, transport, and security options
 */
extern char* SuperRay_GetProtocolList(void);

/* ========== Routing Rules ========== */

/*
 * Create a domain-based routing rule
 * @param domainsJSON: JSON array of domains, e.g. ["geosite:cn", "domain:example.com"]
 * @param outboundTag: Target outbound tag
 * @return JSON with routing rule
 */
extern char* SuperRay_CreateRoutingRuleDomain(const char* domainsJSON, const char* outboundTag);

/*
 * Create an IP-based routing rule
 * @param ipsJSON: JSON array of IPs, e.g. ["geoip:cn", "0.0.0.0/8"]
 * @param outboundTag: Target outbound tag
 * @return JSON with routing rule
 */
extern char* SuperRay_CreateRoutingRuleIP(const char* ipsJSON, const char* outboundTag);

/*
 * Create a port-based routing rule
 * @param portRange: Port range string, e.g. "80,443" or "1-1024"
 * @param outboundTag: Target outbound tag
 * @return JSON with routing rule
 */
extern char* SuperRay_CreateRoutingRulePort(const char* portRange, const char* outboundTag);

/* ========== TUN Device Functions ========== */

/*
 * Create a TUN inbound configuration
 * @param tag: Inbound tag name
 * @param addressesJSON: JSON array of addresses, e.g. ["10.0.0.1/24", "fd00::1/64"]
 * @param mtu: MTU size (default: 1500)
 * @return JSON with inbound configuration
 */
extern char* SuperRay_CreateTUNInbound(const char* tag, const char* addressesJSON, int mtu);

/*
 * Create a TUN inbound with full options
 * @param tag: Inbound tag name
 * @param name: TUN device name (empty for auto)
 * @param addressesJSON: JSON array of addresses
 * @param mtu: MTU size
 * @param autoRoute: Auto configure routing (1=true, 0=false)
 * @return JSON with inbound configuration
 */
extern char* SuperRay_CreateTUNInboundFull(const char* tag, const char* name, const char* addressesJSON, int mtu, int autoRoute);

/*
 * Create a TUN device (gvisor netstack)
 * @param configJSON: JSON config with tag, addresses, mtu
 * @return JSON with device info
 */
extern char* SuperRay_CreateTUNDevice(const char* configJSON);

/*
 * Remove a TUN device
 * @param tag: TUN device tag
 * @return JSON with status
 */
extern char* SuperRay_RemoveTUNDevice(const char* tag);

/*
 * List all TUN devices
 * @return JSON with device list
 */
extern char* SuperRay_ListTUNDevices(void);

/*
 * Write IP packet to TUN device
 * @param tag: TUN device tag
 * @param packetData: Raw IP packet data
 * @param packetLen: Packet length
 * @return JSON with status
 */
extern char* SuperRay_WriteTUNPacket(const char* tag, const char* packetData, int packetLen);

/*
 * Close all TUN devices
 * @return JSON with status
 */
extern char* SuperRay_CloseAllTUNDevices(void);

/*
 * Create TUN device from file descriptor (for mobile platforms)
 * On Android: Pass FD from VpnService.Builder.establish()
 * On iOS: Pass FD from NEPacketTunnelProvider
 * @param fd: File descriptor from platform VPN API
 * @param configJSON: JSON config with mtu, addresses, tag
 * @return JSON with device info
 */
extern char* SuperRay_CreateTUNFromFD(int fd, const char* configJSON);

/*
 * Set callback handler for TUN connections (platform-specific)
 * @param tag: TUN device tag
 * @param callbackID: Callback identifier
 * @return JSON with status
 */
extern char* SuperRay_SetTUNHandler(const char* tag, int callbackID);

/*
 * Get TUN device information
 * @param tag: TUN device tag
 * @return JSON with device info
 */
extern char* SuperRay_GetTUNInfo(const char* tag);

/* ========== Callback-based TUN API for NEPacketTunnelFlow ========== */

/*
 * Create a callback-based TUN device for NEPacketTunnelFlow integration
 * Use this mode when packets are received via NEPacketTunnelFlow.readPackets()
 * and sent via NEPacketTunnelFlow.writePackets()
 *
 * @param configJSON: JSON config with tag, addresses, mtu
 * @return JSON with device info
 *
 * Example usage flow:
 * 1. SuperRay_CreateCallbackTUN() - create the TUN device
 * 2. SuperRay_Run() - start Xray with appropriate config
 * 3. In NEPacketTunnelFlow.readPackets() callback:
 *    - Call SuperRay_EnqueueTUNPacket() for each received packet
 * 4. SuperRay_StopCallbackTUN() - stop when done
 */
extern char* SuperRay_CreateCallbackTUN(const char* configJSON);

/*
 * Enqueue a packet into the callback TUN device
 * Call this from NEPacketTunnelFlow.readPackets() handler
 *
 * @param tag: TUN device tag
 * @param packetData: Raw IP packet data
 * @param packetLen: Packet length
 * @return JSON with bytes count
 */
extern char* SuperRay_EnqueueTUNPacket(const char* tag, const char* packetData, int packetLen);

/*
 * Start the callback TUN device processing
 * @param tag: TUN device tag
 * @return JSON with status
 */
extern char* SuperRay_StartCallbackTUN(const char* tag);

/*
 * Stop and remove a callback TUN device
 * @param tag: TUN device tag
 * @return JSON with status
 */
extern char* SuperRay_StopCallbackTUN(const char* tag);

/*
 * Get callback TUN device information
 * @param tag: TUN device tag
 * @return JSON with device info including running status
 */
extern char* SuperRay_GetCallbackTUNInfo(const char* tag);

/*
 * List all callback TUN devices
 * @return JSON with device list
 */
extern char* SuperRay_ListCallbackTUNs(void);

/*
 * Close all callback TUN devices
 * @return JSON with status
 */
extern char* SuperRay_CloseAllCallbackTUNs(void);

/*
 * Set XrayDialer for callback TUN device
 * This connects the TUN device to an Xray instance for packet forwarding
 * @param tunTag: Callback TUN device tag
 * @param instanceID: Xray instance ID (from SuperRay_Run)
 * @param outboundTag: Xray outbound tag to use (e.g., "proxy"), empty for default
 * @return JSON with success status
 */
extern char* SuperRay_SetCallbackTUNDialer(const char* tunTag, const char* instanceID, const char* outboundTag);

/*
 * Create callback TUN with XrayDialer in one step
 * @param configJSON: TUN config {"tag":"tun0","addresses":["10.0.0.1/24"],"mtu":1500}
 * @param instanceID: Xray instance ID
 * @param outboundTag: Xray outbound tag (empty for "proxy")
 * @return JSON with TUN device info
 */
extern char* SuperRay_CreateCallbackTUNWithDialer(const char* configJSON, const char* instanceID, const char* outboundTag);

/* ========== TUN Packet Output API ========== */

/*
 * Packet output callback function type
 * Called when a packet is ready to be sent to NEPacketTunnelFlow.writePackets()
 *
 * @param data: Packet data pointer
 * @param dataLen: Packet length
 * @param family: AF_INET (2) for IPv4, AF_INET6 (30 on Darwin) for IPv6
 * @param userData: User data passed to SetTUNPacketCallback
 */
typedef void (*SuperRay_PacketOutputCallback)(const void* data, int dataLen, int family, void* userData);

/*
 * Set packet output callback for TUN device
 * When gVisor has a packet ready to send, this callback will be called
 * Use this to send packets back to NEPacketTunnelFlow.writePackets()
 *
 * @param tag: TUN device tag
 * @param callback: C function pointer to receive packets
 * @param userData: User data passed to callback
 * @return JSON with status
 *
 * Example Swift usage:
 *   let callback: @convention(c) (UnsafeRawPointer?, Int32, Int32, UnsafeMutableRawPointer?) -> Void = { data, len, family, _ in
 *       guard let data = data else { return }
 *       let packet = Data(bytes: data, count: Int(len))
 *       let proto = NSNumber(value: family == 2 ? AF_INET : AF_INET6)
 *       packetFlow.writePackets([packet], withProtocols: [proto])
 *   }
 *   SuperRay_SetTUNPacketCallback(tag, callback, nil)
 */
extern char* SuperRay_SetTUNPacketCallback(const char* tag, void* callback, void* userData);

/*
 * Read a packet from TUN output buffer (polling mode, non-blocking)
 * Alternative to SetTUNPacketCallback for applications that prefer polling
 *
 * @param tag: TUN device tag
 * @param buffer: Buffer to receive packet data
 * @param bufferLen: Buffer size (should be >= MTU)
 * @return Number of bytes read, 0 if no packet available, -1 on error
 */
extern int SuperRay_ReadTUNPacket(const char* tag, void* buffer, int bufferLen);

/*
 * Read a packet from TUN output buffer with IP family (polling mode, non-blocking)
 *
 * @param tag: TUN device tag
 * @param buffer: Buffer to receive packet data
 * @param bufferLen: Buffer size
 * @param family: Output parameter for IP family (2=IPv4, 30=IPv6 on Darwin)
 * @return Number of bytes read, 0 if no packet available, -1 on error
 */
extern int SuperRay_ReadTUNPacketWithFamily(const char* tag, void* buffer, int bufferLen, int* family);

/* ========== Traffic Statistics ========== */

/*
 * Get traffic statistics
 * @return JSON with upload, download bytes and connection count
 */
extern char* SuperRay_GetTrafficStats(void);

/*
 * Reset traffic statistics
 * @return JSON with status
 */
extern char* SuperRay_ResetTrafficStats(void);

/*
 * Get active connections
 * @return JSON with connection list
 */
extern char* SuperRay_GetConnections(void);

/*
 * Get active connection count
 * @return Number of active connections
 */
extern int SuperRay_GetConnectionCount(void);

/* ========== Xray Core Stats (Direct Function Export, No gRPC) ========== */

/*
 * Get Xray core traffic statistics from all running instances
 * @return JSON: {"success":true,"data":{"uplink":123,"downlink":456,"uplink_rate":100.5,"downlink_rate":200.3,"users":{},"inbounds":{},"outbounds":{}}}
 * Note: Requires "stats":{} in Xray config to enable statistics
 */
extern char* SuperRay_GetXrayStats(void);

/*
 * Get Xray core stats for a specific instance
 * @param instanceID: The instance ID returned by SuperRay_Run
 * @return JSON with stats for the specified instance
 */
extern char* SuperRay_GetXrayStatsForInstance(const char* instanceID);

/*
 * Reset all Xray stats counters
 * @return JSON with status
 */
extern char* SuperRay_ResetXrayStats(void);

/*
 * Get current upload/download speed
 * @return JSON: {"success":true,"data":{"uplink_rate":1234.5,"downlink_rate":5678.9,"uplink_kbps":1.2,"downlink_kbps":5.5,"uplink_mbps":0.001,"downlink_mbps":0.005}}
 * Note: Rate is calculated based on the time since the last call to this function
 */
extern char* SuperRay_GetCurrentSpeed(void);

/* ========== Subscription Management ========== */

/*
 * Add a subscription
 * @param name: Subscription name
 * @param url: Subscription URL
 * @return JSON with status
 */
extern char* SuperRay_AddSubscription(const char* name, const char* url);

/*
 * Remove a subscription
 * @param name: Subscription name
 * @return JSON with status
 */
extern char* SuperRay_RemoveSubscription(const char* name);

/*
 * Update a subscription (fetch and parse)
 * @param name: Subscription name
 * @return JSON with subscription info and servers
 */
extern char* SuperRay_UpdateSubscription(const char* name);

/*
 * Update all subscriptions
 * @return JSON with results for each subscription
 */
extern char* SuperRay_UpdateAllSubscriptions(void);

/*
 * Get subscription info
 * @param name: Subscription name
 * @return JSON with subscription info
 */
extern char* SuperRay_GetSubscription(const char* name);

/*
 * List all subscriptions
 * @return JSON with subscription names
 */
extern char* SuperRay_ListSubscriptions(void);

/*
 * Get all servers from all subscriptions
 * @return JSON with server list
 */
extern char* SuperRay_GetAllServers(void);

/*
 * Export subscription as JSON
 * @param name: Subscription name
 * @return JSON with subscription data
 */
extern char* SuperRay_ExportSubscription(const char* name);

/*
 * Import subscription from JSON
 * @param jsonData: Subscription JSON data
 * @return JSON with status
 */
extern char* SuperRay_ImportSubscription(const char* jsonData);

/* ========== Logging ========== */

/*
 * Set log level
 * @param level: Log level ("debug", "info", "warning", "error", "none")
 * @return JSON with status
 */
extern char* SuperRay_SetLogLevel(const char* level);

/*
 * Get current log level
 * @return JSON with current level
 */
extern char* SuperRay_GetLogLevel(void);

/*
 * Get recent log entries
 * @param count: Number of entries to retrieve
 * @return JSON with log entries
 */
extern char* SuperRay_GetRecentLogs(int count);

/*
 * Clear log buffer
 * @return JSON with status
 */
extern char* SuperRay_ClearLogs(void);

/*
 * Write a log entry
 * @param level: Log level
 * @param tag: Log tag
 * @param message: Log message
 * @return JSON with status
 */
extern char* SuperRay_Log(const char* level, const char* tag, const char* message);

/* ========== Speed Test / Latency ========== */

/*
 * TCP ping to test latency
 * @param address: Server address
 * @param port: Server port
 * @param timeoutMs: Timeout in milliseconds
 * @return JSON with latency result
 */
extern char* SuperRay_TCPPing(const char* address, int port, int timeoutMs);

/*
 * TCP ping multiple times
 * @param address: Server address
 * @param port: Server port
 * @param count: Number of pings
 * @param timeoutMs: Timeout per ping
 * @return JSON with average, min, max latency
 */
extern char* SuperRay_TCPPingMultiple(const char* address, int port, int count, int timeoutMs);

/*
 * Batch latency test for multiple servers
 * @param serversJSON: JSON array of servers [{address, port, name}]
 * @param concurrent: Max concurrent tests
 * @param count: Pings per server
 * @param timeoutMs: Timeout per ping
 * @return JSON with sorted results
 */
extern char* SuperRay_BatchLatencyTest(const char* serversJSON, int concurrent, int count, int timeoutMs);

/*
 * HTTP ping through proxy
 * @param url: URL to ping
 * @param proxyAddr: Proxy address (host:port) or empty
 * @param timeoutMs: Timeout in milliseconds
 * @return JSON with latency and status code
 */
extern char* SuperRay_HTTPPing(const char* url, const char* proxyAddr, int timeoutMs);

/*
 * Run download speed test
 * @param downloadURL: URL to download from
 * @param proxyAddr: Proxy address (host:port) or empty
 * @param durationSec: Test duration in seconds
 * @return JSON with download speed in Mbps
 */
extern char* SuperRay_SpeedTest(const char* downloadURL, const char* proxyAddr, int durationSec);

/*
 * Test all servers in a subscription
 * @param subscriptionName: Name of subscription
 * @param concurrent: Max concurrent tests
 * @param timeoutMs: Timeout per test
 * @return JSON with sorted results
 */
extern char* SuperRay_TestSubscriptionServers(const char* subscriptionName, int concurrent, int timeoutMs);

/* ========== Auto Failover ========== */

/*
 * Setup automatic failover
 * @param serversJSON: JSON array of servers
 * @param checkIntervalSec: Health check interval in seconds
 * @param failThreshold: Consecutive failures to trigger switch
 * @param latencyLimitMs: Max acceptable latency (0 = no limit)
 * @return JSON with status
 */
extern char* SuperRay_SetupFailover(const char* serversJSON, int checkIntervalSec, int failThreshold, int latencyLimitMs);

/*
 * Start failover monitoring
 * @return JSON with status
 */
extern char* SuperRay_StartFailover(void);

/*
 * Stop failover monitoring
 * @return JSON with status
 */
extern char* SuperRay_StopFailover(void);

/*
 * Get current active server
 * @return JSON with server info
 */
extern char* SuperRay_GetCurrentServer(void);

/*
 * Manually switch to a server
 * @param index: Server index
 * @return JSON with status
 */
extern char* SuperRay_SwitchServer(int index);

/* ========== iOS Memory Optimization ========== */

/*
 * Initialize iOS memory optimizations
 * Should be called as early as possible in the iOS app lifecycle
 * On non-iOS platforms, this is a no-op
 *
 * @param configJSON: JSON configuration string, or NULL for defaults
 *   {
 *     "memory_limit_mb": 12,      // Soft memory limit (8-14 MB for iOS)
 *     "max_procs": 2,             // GOMAXPROCS (1-4)
 *     "gc_percent": 50,           // GOGC percentage (lower = more frequent GC)
 *     "gc_interval_seconds": 30   // Periodic GC interval (0 to disable)
 *   }
 * @return JSON with previous GOMAXPROCS and applied config
 */
extern char* SuperRay_InitIOSMemory(const char* configJSON);

/*
 * Initialize iOS memory with default settings (12MB limit, GOMAXPROCS=2)
 * @return JSON with configuration
 */
extern char* SuperRay_InitIOSMemoryDefault(void);

/*
 * Initialize iOS memory with aggressive settings (8MB limit, GOMAXPROCS=1)
 * Use for very constrained memory environments
 * @return JSON with configuration
 */
extern char* SuperRay_InitIOSMemoryAggressive(void);

/*
 * Get current memory usage statistics
 * Works on all platforms
 * @return JSON with memory stats (alloc_mb, heap_mb, num_gc, etc.)
 */
extern char* SuperRay_GetMemoryStats(void);

/*
 * Force immediate garbage collection
 * @return JSON with memory stats after GC
 */
extern char* SuperRay_ForceGC(void);

/*
 * Handle iOS memory warning
 * Should be called when iOS sends didReceiveMemoryWarning
 * Aggressively frees memory and temporarily increases GC frequency
 * @return JSON with handled status and memory stats
 */
extern char* SuperRay_HandleMemoryWarning(void);

/*
 * Check if memory usage is approaching the limit
 * Returns true if heap usage is above 80% of configured limit
 * Only meaningful on iOS; always returns false on other platforms
 * @return JSON with memory_pressure boolean
 */
extern char* SuperRay_IsMemoryPressure(void);

/*
 * Stop periodic GC goroutine if running
 * Should be called when shutting down
 * @return JSON with stopped status
 */
extern char* SuperRay_StopPeriodicGC(void);

/* ========== System TUN (Desktop Platforms) ========== */

/*
 * Create a system-level TUN device (requires root/admin)
 * @param configJSON: {"tag":"tun0","name":"","mtu":1500,"addresses":["10.255.0.1/24"]}
 * @return JSON with tag, name, mtu, status
 */
extern char* SuperRay_CreateSystemTUN(const char* configJSON);

/*
 * Start TUN stack connected to Xray instance
 * @param tag: TUN device tag
 * @param instanceID: Xray instance ID
 * @param outboundTag: Outbound tag to use (default: "proxy")
 * @return JSON with status
 */
extern char* SuperRay_StartSystemTUNStack(const char* tag, const char* instanceID, const char* outboundTag);

/*
 * Setup system routes for TUN
 * @param tag: TUN device tag
 * @param serverAddress: VPN server address (to exclude from TUN)
 * @return JSON with status
 */
extern char* SuperRay_SetupRoutes(const char* tag, const char* serverAddress);

/*
 * Cleanup system routes
 * @param tag: TUN device tag
 * @return JSON with status
 */
extern char* SuperRay_CleanupRoutes(const char* tag);

/*
 * Close a system TUN device
 * @param tag: TUN device tag
 * @return JSON with status
 */
extern char* SuperRay_CloseSystemTUN(const char* tag);

/*
 * Close all system TUN devices
 * @return JSON with status
 */
extern char* SuperRay_CloseAllSystemTUNs(void);

/* ========== Memory Management ========== */

/*
 * Free memory allocated by SuperRay functions
 * Must be called for every returned char* to prevent memory leaks
 * @param ptr: Pointer returned by SuperRay functions
 */
extern void SuperRay_Free(char* ptr);

/*
 * Free raw bytes allocated by SuperRay
 * @param ptr: Pointer to free
 */
extern void SuperRay_FreeBytes(void* ptr);

/* ==========================================================================
 * libXray Compatibility API
 * ==========================================================================
 * These functions provide API compatibility with libXray for easy migration.
 * They use base64-encoded JSON for requests and responses, matching libXray's format.
 *
 * Response format (base64-encoded JSON):
 * {
 *   "success": true|false,
 *   "data": ...,
 *   "error": "..."
 * }
 */

/* ---------- Xray Instance Control ---------- */

/*
 * Run Xray from config file path (libXray compatible)
 * @param base64Text: Base64-encoded JSON: {"datDir":"...", "configPath":"..."}
 * @return Base64-encoded JSON response
 */
extern char* LibXray_RunXray(const char* base64Text);

/*
 * Run Xray from JSON config (libXray compatible)
 * @param base64Text: Base64-encoded JSON: {"datDir":"...", "configJSON":"..."}
 * @return Base64-encoded JSON response
 */
extern char* LibXray_RunXrayFromJSON(const char* base64Text);

/*
 * Stop running Xray instance (libXray compatible)
 * @return Base64-encoded JSON response
 */
extern char* LibXray_StopXray(void);

/*
 * Get Xray running state (libXray compatible)
 * @return 1 if running, 0 if not running
 */
extern int LibXray_GetXrayState(void);

/*
 * Test Xray config without starting (libXray compatible)
 * @param base64Text: Base64-encoded JSON: {"datDir":"...", "configPath":"..."}
 * @return Base64-encoded JSON response
 */
extern char* LibXray_TestXray(const char* base64Text);

/*
 * Get Xray version (libXray compatible)
 * @return Base64-encoded JSON response with version string
 */
extern char* LibXray_XrayVersion(void);

/* ---------- Network Functions ---------- */

/*
 * Ping through Xray config (libXray compatible)
 * @param base64Text: Base64-encoded JSON: {"datDir":"...", "configPath":"...", "timeout":5000, "url":"...", "proxy":"..."}
 * @return Base64-encoded JSON response with latency in ms
 */
extern char* LibXray_Ping(const char* base64Text);

/*
 * Query traffic stats (libXray compatible)
 * @param base64Text: Base64-encoded server address (for gRPC stats API)
 * @return Base64-encoded JSON response with stats
 */
extern char* LibXray_QueryStats(const char* base64Text);

/*
 * Get free ports (libXray compatible)
 * @param count: Number of ports to find
 * @return Base64-encoded JSON: {"ports":[...]}
 */
extern char* LibXray_GetFreePorts(int count);

/* ---------- Share Link Functions ---------- */

/*
 * Convert share links to Xray JSON config (libXray compatible)
 * @param base64Text: Base64-encoded share links (one per line)
 * @return Base64-encoded JSON Xray config
 */
extern char* LibXray_ConvertShareLinksToXrayJson(const char* base64Text);

/*
 * Convert Xray JSON config to share links (libXray compatible)
 * @param base64Text: Base64-encoded Xray JSON config
 * @return Base64-encoded share links
 */
extern char* LibXray_ConvertXrayJsonToShareLinks(const char* base64Text);

/* ---------- Geo Data Functions ---------- */

/*
 * Count entries in geo data file (libXray compatible)
 * @param base64Text: Base64-encoded JSON: {"datDir":"...", "name":"...", "geoType":"ip|site"}
 * @return Base64-encoded JSON response
 */
extern char* LibXray_CountGeoData(const char* base64Text);

/*
 * Read geo file references from config (libXray compatible)
 * @param base64Text: Base64-encoded Xray JSON config
 * @return Base64-encoded JSON: {"domain":[...], "ip":[...]}
 */
extern char* LibXray_ReadGeoFiles(const char* base64Text);

/* ---------- DNS Functions ---------- */

/*
 * Initialize DNS settings (libXray compatible)
 * @param base64Text: Base64-encoded JSON: {"dns":"...", "deviceName":"..."}
 * @return Base64-encoded JSON response
 */
extern char* LibXray_InitDns(const char* base64Text);

/*
 * Reset DNS to system default (libXray compatible)
 * @return Base64-encoded JSON response
 */
extern char* LibXray_ResetDns(void);

/* ---------- Request Helper Functions ---------- */

/*
 * Create a RunXray request (libXray compatible)
 * @param datDir: Geo data directory
 * @param configPath: Config file path
 * @return Base64-encoded JSON request
 */
extern char* LibXray_NewXrayRunRequest(const char* datDir, const char* configPath);

/*
 * Create a RunXrayFromJSON request (libXray compatible)
 * @param datDir: Geo data directory
 * @param configJSON: Xray JSON config
 * @return Base64-encoded JSON request
 */
extern char* LibXray_NewXrayRunFromJSONRequest(const char* datDir, const char* configJSON);

/*
 * Create an InitDns request (libXray compatible)
 * @param dns: DNS server
 * @param deviceName: Device name
 * @return Base64-encoded JSON request
 */
extern char* LibXray_NewInitDnsRequest(const char* dns, const char* deviceName);

/*
 * Create a Ping request (libXray compatible)
 * @param datDir: Geo data directory
 * @param configPath: Config file path
 * @param timeout: Timeout in milliseconds
 * @param url: URL to ping
 * @param proxy: Proxy address (host:port)
 * @return Base64-encoded JSON request
 */
extern char* LibXray_NewPingRequest(const char* datDir, const char* configPath, int timeout, const char* url, const char* proxy);

#ifdef __cplusplus
}
#endif

#endif /* SUPERRAY_H */
