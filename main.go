package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"superray-tui/pkg/superray"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App represents the TUI application
type App struct {
	app         *tview.Application
	pages       *tview.Pages
	mainFlex    *tview.Flex
	statusView  *tview.TextView
	speedView   *tview.TextView
	serverList  *tview.List
	connList    *tview.Table
	logView     *tview.TextView
	helpView    *tview.TextView

	// State
	mu              sync.RWMutex
	servers         []*superray.Server
	currentServer   *superray.Server
	selectedIndex   int
	instanceID      string
	isConnected     bool
	isConnecting    bool
	totalUpload     int64
	totalDownload   int64
	uploadSpeed     float64
	downloadSpeed   float64
	lastUpdateTime  time.Time
	lastStats       *superray.TrafficStats

	// Traffic history for chart
	trafficHistory []TrafficPoint
	historyMaxLen  int

	// Screen for resize
	screen tcell.Screen

	// Config
	subscriptionURL string
	localPort       int
	geoPath         string
	accessLogPath   string
	errorLogPath    string
	directCountries []string
	tunMode         bool

	// GeoIP info for selected server
	serverGeoInfo *GeoIPInfo
}

// TrafficPoint represents a point in traffic history
type TrafficPoint struct {
	Time     time.Time
	Upload   int64
	Download int64
}

// GeoIPInfo contains detailed information about an IP address
type GeoIPInfo struct {
	IP          string `json:"ip"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	Region      string `json:"region"`
	RegionName  string `json:"regionName"`
	City        string `json:"city"`
	ISP         string `json:"isp"`
	Org         string `json:"org"`
	AS          string `json:"as"`
	ASName      string `json:"asname"`
	Query       string `json:"query"`
	Status      string `json:"status"`
}

// IP info cache
var (
	ipCache     = make(map[string]*GeoIPInfo)
	ipCacheMu   sync.RWMutex
	ipCacheTime = make(map[string]time.Time)
	cacheTTL    = 30 * time.Minute
)

// Global log path for panic handlers
var globalErrorLogPath = "error.log"

func main() {
	app := NewApp()
	app.loadEnvConfig()

	globalErrorLogPath = app.errorLogPath

	// Write startup log
	f, _ := os.OpenFile(globalErrorLogPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "=== SuperRay TUI Started ===\n")
		fmt.Fprintf(f, "Time: %s\n\n", time.Now().Format(time.RFC3339))
		f.Close()
	}

	// Catch panics
	defer func() {
		if r := recover(); r != nil {
			f, _ := os.OpenFile(globalErrorLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if f != nil {
				fmt.Fprintf(f, "=== MAIN PANIC ===\n")
				fmt.Fprintf(f, "Time: %s\n", time.Now().Format(time.RFC3339))
				fmt.Fprintf(f, "Panic: %v\n\n", r)
				fmt.Fprintf(f, "Stack Trace:\n%s\n", debug.Stack())
				f.Close()
			}
			os.Exit(1)
		}
	}()

	if err := app.Run(); err != nil {
		f, _ := os.OpenFile(globalErrorLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if f != nil {
			fmt.Fprintf(f, "Error: %v\n", err)
			f.Close()
		}
		os.Exit(1)
	}
}

func NewApp() *App {
	return &App{
		localPort:       10808,
		geoPath:         "./geoip",
		accessLogPath:   "access.log",
		errorLogPath:    "error.log",
		directCountries: []string{"cn"},
		lastUpdateTime:  time.Now(),
		historyMaxLen:   300,
		trafficHistory:  make([]TrafficPoint, 0, 300),
	}
}

// safeGo runs a function in a goroutine with panic recovery
func safeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				f, _ := os.OpenFile(globalErrorLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
				if f != nil {
					fmt.Fprintf(f, "=== Goroutine Panic ===\n")
					fmt.Fprintf(f, "Time: %s\n", time.Now().Format(time.RFC3339))
					fmt.Fprintf(f, "Panic: %v\n\n", r)
					fmt.Fprintf(f, "Stack Trace:\n%s\n\n", debug.Stack())
					f.Close()
				}
			}
		}()
		fn()
	}()
}

// extractLatency extracts latency value from server name like "Server [50ms]"
// Returns -1 if no valid latency found
func extractLatency(name string) int {
	// Find pattern like [123ms]
	start := strings.LastIndex(name, "[")
	end := strings.LastIndex(name, "ms]")
	if start == -1 || end == -1 || end <= start {
		return -1
	}
	latencyStr := name[start+1 : end]
	if latencyStr == "timeout" {
		return -1
	}
	var latency int
	_, err := fmt.Sscanf(latencyStr, "%d", &latency)
	if err != nil {
		return -1
	}
	return latency
}

// loadEnvConfig loads configuration from .env file
func (a *App) loadEnvConfig() {
	envPaths := []string{
		".env",
		filepath.Join(filepath.Dir(os.Args[0]), ".env"),
	}

	for _, path := range envPaths {
		if err := a.loadEnvFile(path); err == nil {
			break
		}
	}

	// Also check environment variables
	if url := os.Getenv("SUPERRAY_SUB_URL"); url != "" {
		a.subscriptionURL = url
	}
	if port := os.Getenv("SUPERRAY_LOCAL_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &a.localPort)
	}
	if geoPath := os.Getenv("SUPERRAY_GEO_PATH"); geoPath != "" {
		a.geoPath = geoPath
	}
	if directCountries := os.Getenv("DIRECT_COUNTRIES"); directCountries != "" {
		a.directCountries = nil
		for _, c := range strings.Split(directCountries, ",") {
			c = strings.TrimSpace(c)
			if c != "" {
				a.directCountries = append(a.directCountries, strings.ToLower(c))
			}
		}
	}

	// Set geo asset directory
	a.setupGeoPath()
}

// setupGeoPath configures the geo data path
func (a *App) setupGeoPath() {
	geoPath := a.geoPath
	if !filepath.IsAbs(geoPath) {
		if absPath, err := filepath.Abs(geoPath); err == nil {
			geoPath = absPath
		}
	}
	superray.SetAssetDir(geoPath)
}

func (a *App) loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'")

		switch key {
		case "SUPERRAY_SUB_URL":
			a.subscriptionURL = value
		case "SUPERRAY_LOCAL_PORT":
			fmt.Sscanf(value, "%d", &a.localPort)
		case "SUPERRAY_GEO_PATH":
			a.geoPath = value
		case "ACCESS_LOG":
			a.accessLogPath = value
		case "ERROR_LOG":
			a.errorLogPath = value
		case "DIRECT_COUNTRIES":
			a.directCountries = nil
			for _, c := range strings.Split(value, ",") {
				c = strings.TrimSpace(c)
				if c != "" {
					a.directCountries = append(a.directCountries, strings.ToLower(c))
				}
			}
		}
	}

	return scanner.Err()
}

// maskAddress masks server address for privacy
func maskAddress(addr string) string {
	parts := strings.Split(addr, ".")
	if len(parts) == 4 {
		isIPv4 := true
		for _, p := range parts {
			if _, err := fmt.Sscanf(p, "%d", new(int)); err != nil {
				isIPv4 = false
				break
			}
		}
		if isIPv4 {
			return fmt.Sprintf("%s.%s.*.*", parts[0], parts[1])
		}
	}

	if strings.Contains(addr, ":") {
		ipv6Parts := strings.Split(addr, ":")
		if len(ipv6Parts) > 2 {
			return fmt.Sprintf("%s:%s:***", ipv6Parts[0], ipv6Parts[1])
		}
	}

	if len(parts) >= 2 {
		if len(parts) >= 3 {
			return strings.Join(parts[:len(parts)-2], ".") + ".***"
		} else if len(parts) == 2 {
			return parts[0] + ".***"
		}
	}

	if len(addr) > 6 {
		return addr[:len(addr)-4] + "****"
	}
	return addr
}

// maskIPAddress masks the last two segments of an IP address
// e.g., 192.168.1.100 -> 192.168.*.*
func maskIPAddress(ip string) string {
	// IPv4
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return fmt.Sprintf("%s.%s.*.*", parts[0], parts[1])
	}

	// IPv6
	if strings.Contains(ip, ":") {
		ipv6Parts := strings.Split(ip, ":")
		if len(ipv6Parts) >= 2 {
			// Mask the last two segments
			if len(ipv6Parts) >= 2 {
				ipv6Parts[len(ipv6Parts)-1] = "****"
				ipv6Parts[len(ipv6Parts)-2] = "****"
				return strings.Join(ipv6Parts, ":")
			}
		}
	}

	return ip
}

func (a *App) Run() error {
	a.app = tview.NewApplication()
	a.app.EnableMouse(false)
	a.setupUI()
	a.startUpdateLoop()

	a.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		a.screen = screen
		return false
	})

	safeGo(func() {
		time.Sleep(100 * time.Millisecond)
		a.showVersion()
		a.log("[green]SuperRay TUI Client[white]")
		a.log("Press [yellow]s[white] to set subscription URL")
		a.log("Press [yellow]r[white] to refresh servers")
		a.log("Press [yellow]u[white] to toggle TUN mode (requires admin)")
		a.log("Press [yellow]c[white] or [yellow]Enter[white] to connect")

		if a.subscriptionURL != "" {
			a.log("[green]Found subscription URL in config, loading...[white]")
			a.loadSubscription()
		}
	})

	return a.app.Run()
}

func (a *App) setupUI() {
	a.createStatusView()
	a.createSpeedView()
	a.createConnList()
	a.createServerList()
	a.createLogView()
	a.createHelpView()

	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.statusView, 14, 0, false).
		AddItem(a.speedView, 8, 0, false).
		AddItem(a.connList, 0, 1, false)

	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.serverList, 22, 0, true).
		AddItem(a.logView, 0, 1, false)

	mainContent := tview.NewFlex().
		AddItem(leftPanel, 0, 1, false).
		AddItem(rightPanel, 0, 1, true)

	a.mainFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainContent, 0, 1, true).
		AddItem(a.helpView, 1, 0, false)

	a.pages = tview.NewPages().
		AddPage("main", a.mainFlex, true, true)

	a.setupKeyBindings()

	a.app.SetRoot(a.pages, true)
	a.app.SetFocus(a.serverList)
}

func (a *App) createStatusView() {
	a.statusView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	a.statusView.SetBorder(true).SetTitle(" SuperRay ")
	a.updateStatusView()
}

func (a *App) createSpeedView() {
	a.speedView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	a.speedView.SetBorder(true).SetTitle(" Speed ")
	a.updateSpeedView()
}

func (a *App) createLogView() {
	a.logView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true).
		SetMaxLines(200)
	a.logView.SetBorder(true).SetTitle(" Logs ")
}

func (a *App) createServerList() {
	a.serverList = tview.NewList().
		SetHighlightFullLine(true).
		SetSelectedBackgroundColor(tcell.ColorDarkBlue).
		SetSelectedTextColor(tcell.ColorWhite).
		SetMainTextColor(tcell.ColorWhite).
		SetSecondaryTextColor(tcell.ColorGray).
		ShowSecondaryText(false)
	a.serverList.SetBorder(true).SetTitle(" Servers [Enter/c: Connect] ")

	a.serverList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		a.mu.Lock()
		a.selectedIndex = index
		var serverAddr string
		if index >= 0 && index < len(a.servers) {
			serverAddr = a.servers[index].Address
		}
		total := len(a.servers)
		a.mu.Unlock()

		// Update title with scroll position
		if total > 0 {
			a.serverList.SetTitle(fmt.Sprintf(" Servers [%d/%d] ", index+1, total))
		}

		// Async lookup GeoIP info
		if serverAddr != "" {
			safeGo(func() { a.lookupServerGeoIP(serverAddr) })
		}
	})

	a.serverList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})
}

func (a *App) createConnList() {
	a.connList = tview.NewTable().
		SetBorders(false).
		SetSelectable(false, false)
	a.connList.SetBorder(true).SetTitle(" Connections ")
	a.updateConnList()
}

func (a *App) createHelpView() {
	a.helpView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	a.helpView.SetText("[yellow]q[white]:Quit [yellow]c[white]:Connect [yellow]d[white]:Disconnect [yellow]r[white]:Load [yellow]s[white]:Sub [yellow]t[white]:Test [yellow]u[white]:TUN [yellow]f[white]:Refresh")
}

func (a *App) setupKeyBindings() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if a.pages.HasPage("modal") {
			if event.Key() == tcell.KeyEscape {
				a.pages.RemovePage("modal")
				return nil
			}
			return event
		}

		switch event.Key() {
		case tcell.KeyEnter:
			safeGo(func() { a.connectSelected() })
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				a.quit()
				return nil
			case 'c', 'C', ' ':
				safeGo(func() { a.connectSelected() })
				return nil
			case 'd', 'D':
				safeGo(func() { a.doDisconnect() })
				return nil
			case 'r', 'R':
				safeGo(func() { a.loadSubscription() })
				return nil
			case 's', 'S':
				a.showSubscriptionDialog()
				return nil
			case 't', 'T':
				safeGo(func() { a.testLatency() })
				return nil
			case 'u', 'U':
				safeGo(func() { a.toggleTunMode() })
				return nil
			case 'f', 'F':
				a.forceRefresh()
				return nil
			}
		}
		return event
	})
}

func (a *App) doDisconnect() {
	a.log("[yellow]Disconnecting...[white]")
	a.disconnect()
}

func (a *App) quit() {
	// Get state
	a.mu.Lock()
	wasConnected := a.isConnected
	tunMode := a.tunMode
	instanceID := a.instanceID
	a.isConnected = false
	a.instanceID = ""
	a.currentServer = nil
	a.mu.Unlock()

	// Cleanup synchronously before stopping
	if tunMode {
		superray.CleanupRoutes("tun0")
		superray.CloseAllSystemTUNs()
	}
	superray.CloseAllCallbackTUNs()
	superray.CloseAllTUNDevices()
	if wasConnected && instanceID != "" {
		superray.DestroyInstance(instanceID)
	}

	// Stop app
	a.app.Stop()
}

func (a *App) startUpdateLoop() {
	safeGo(func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			a.collectAndUpdateStats()

			a.app.QueueUpdateDraw(func() {
				a.renderAllViews()
			})
		}
	})
}

func (a *App) collectAndUpdateStats() {
	a.mu.RLock()
	connected := a.isConnected
	connecting := a.isConnecting
	a.mu.RUnlock()

	if !connected || connecting {
		return
	}

	// Get stats from Xray core (includes inbounds/outbounds)
	stats, err := superray.GetXrayStats()
	if err != nil || stats == nil {
		return
	}

	a.mu.Lock()
	now := time.Now()
	elapsed := now.Sub(a.lastUpdateTime).Seconds()

	// Calculate speed from difference
	if elapsed > 0 && a.lastStats != nil {
		a.uploadSpeed = float64(stats.Uplink-a.lastStats.Uplink) / elapsed
		a.downloadSpeed = float64(stats.Downlink-a.lastStats.Downlink) / elapsed
	}

	// Update totals
	a.totalUpload = stats.Uplink
	a.totalDownload = stats.Downlink

	// Store traffic history
	a.trafficHistory = append(a.trafficHistory, TrafficPoint{
		Time:     now,
		Upload:   stats.Uplink,
		Download: stats.Downlink,
	})

	if len(a.trafficHistory) > a.historyMaxLen {
		a.trafficHistory = a.trafficHistory[1:]
	}

	// Save current stats for next calculation
	a.lastStats = stats
	a.lastUpdateTime = now
	a.mu.Unlock()
}

func (a *App) renderAllViews() {
	a.updateStatusView()
	a.updateSpeedView()
	a.updateConnList()
}

func (a *App) forceRefresh() {
	a.app.QueueUpdateDraw(func() {
		a.updateStatusView()
		a.updateSpeedView()
		a.updateConnList()
		a.updateServerList()
		if a.screen != nil {
			a.screen.Sync()
		}
	})
}

func (a *App) updateStatusView() {
	a.mu.RLock()
	isConnecting := a.isConnecting
	isConnected := a.isConnected
	tunMode := a.tunMode
	totalUpload := a.totalUpload
	totalDownload := a.totalDownload
	currentServer := a.currentServer
	selectedIndex := a.selectedIndex
	servers := make([]*superray.Server, len(a.servers))
	copy(servers, a.servers)
	serverGeoInfo := a.serverGeoInfo
	localPort := a.localPort
	a.mu.RUnlock()

	var status string
	if isConnecting {
		status = "[yellow]◐ Connecting[white]"
	} else if isConnected {
		status = "[green]● Connected[white]"
	} else {
		status = "[red]○ Disconnected[white]"
	}

	var mode string
	if tunMode {
		mode = "[aqua]TUN[-]"
	} else {
		mode = "[yellow]SOCKS5[-]"
	}

	trafficLine := ""
	if isConnected {
		trafficLine = fmt.Sprintf(" Traffic: [green]↑[white]%s [blue]↓[white]%s [yellow]Σ[white]%s",
			formatBytes(totalUpload),
			formatBytes(totalDownload),
			formatBytes(totalUpload+totalDownload))
	}

	serverSection := ""
	var displayServer *superray.Server

	if isConnected && currentServer != nil {
		displayServer = currentServer
	} else if selectedIndex >= 0 && len(servers) > 0 && selectedIndex < len(servers) {
		displayServer = servers[selectedIndex]
	}

	if displayServer != nil {
		connMark := ""
		if isConnected && currentServer != nil &&
			displayServer.Address == currentServer.Address &&
			displayServer.Port == currentServer.Port {
			connMark = " [green]✓[white]"
		}

		serverSection = fmt.Sprintf(" Server: [green]%s[-]%s\n Addr: %s:%d [yellow]%s[-]",
			displayServer.Name, connMark,
			maskAddress(displayServer.Address), displayServer.Port,
			strings.ToUpper(displayServer.Protocol))
	} else {
		serverSection = " [darkgray]No server selected[-]"
	}

	// GeoIP info line
	geoLine := ""
	if serverGeoInfo != nil && serverGeoInfo.Status == "success" {
		var geoParts []string
		// Masked IP address
		if serverGeoInfo.Query != "" {
			geoParts = append(geoParts, fmt.Sprintf("IP:[white]%s[-]", maskIPAddress(serverGeoInfo.Query)))
		}
		// ASN
		if serverGeoInfo.AS != "" {
			geoParts = append(geoParts, fmt.Sprintf("ASN:[yellow]%s[-]", serverGeoInfo.AS))
		}
		// Country and City
		location := ""
		if serverGeoInfo.Country != "" {
			location = serverGeoInfo.Country
		}
		if serverGeoInfo.City != "" && serverGeoInfo.City != serverGeoInfo.Country {
			location += " " + serverGeoInfo.City
		}
		if location != "" {
			geoParts = append(geoParts, fmt.Sprintf("[lime]%s[-]", location))
		}
		// Organization/ISP
		if org := serverGeoInfo.FormatOrg(); org != "" {
			geoParts = append(geoParts, fmt.Sprintf("[aqua]%s[-]", org))
		}
		if len(geoParts) > 0 {
			geoLine = fmt.Sprintf("\n GeoIP: %s", strings.Join(geoParts, " | "))
		}
	}

	logo := `[green] ___                  ___
/ __|_  _ _ __  ___ _| _ \__ _ _  _
\__ \ || | '_ \/ -_) |   / _' | || |
|___/\_,_| .__/\___|_|_|_\__,_|\_, |
         |_|                   |__/[-]`

	text := fmt.Sprintf("%s\n\n %s | %s | Port:[lime]%d[-]%s\n%s%s",
		logo, status, mode, localPort, trafficLine, serverSection, geoLine)

	a.statusView.SetText(text)
}

func (a *App) updateSpeedView() {
	a.mu.RLock()
	uploadSpeed := a.uploadSpeed
	downloadSpeed := a.downloadSpeed
	trafficHistory := make([]TrafficPoint, len(a.trafficHistory))
	copy(trafficHistory, a.trafficHistory)
	a.mu.RUnlock()

	upSpeed := formatSpeed(uploadSpeed)
	downSpeed := formatSpeed(downloadSpeed)

	_, _, width, height := a.speedView.GetInnerRect()
	chartWidth := width - 2
	if chartWidth < 10 {
		chartWidth = 10
	}
	chartHeight := height - 2
	if chartHeight < 3 {
		chartHeight = 3
	}

	chart := createTrafficChart(trafficHistory, chartWidth, chartHeight)
	text := fmt.Sprintf(" [green]↑[-] %s  [blue]↓[-] %s\n%s", upSpeed, downSpeed, chart)

	a.speedView.SetText(text)
}

func createTrafficChart(trafficHistory []TrafficPoint, width, height int) string {
	if width <= 0 {
		width = 10
	}
	if height <= 0 {
		height = 3
	}

	if len(trafficHistory) == 0 {
		lines := make([]string, height)
		for i := 0; i < height; i++ {
			lines[i] = " [darkgray]" + strings.Repeat("─", width) + "[-]"
		}
		return strings.Join(lines, "\n")
	}

	downloads := make([]int64, len(trafficHistory))
	for i, p := range trafficHistory {
		downloads[i] = p.Download
	}

	var maxVal int64 = 1
	for _, v := range downloads {
		if v > maxVal {
			maxVal = v
		}
	}

	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	dataLen := len(downloads)

	for col := 0; col < width; col++ {
		idx := col * dataLen / width
		if idx >= dataLen {
			idx = dataLen - 1
		}
		if idx < 0 {
			continue
		}

		val := downloads[idx]
		barHeight := int(float64(val) / float64(maxVal) * float64(height-1))
		if barHeight >= height {
			barHeight = height - 1
		}
		if barHeight < 0 {
			barHeight = 0
		}

		for h := 0; h <= barHeight; h++ {
			row := height - 1 - h
			if row >= 0 && row < height && col >= 0 && col < width {
				grid[row][col] = '█'
			}
		}
	}

	lines := make([]string, height)
	for i := 0; i < height; i++ {
		line := " [green]"
		for j := 0; j < width; j++ {
			if grid[i][j] == ' ' {
				line += "·"
			} else {
				line += string(grid[i][j])
			}
		}
		line += "[-]"
		lines[i] = line
	}

	maxStr := formatBytes(maxVal)
	dotsCount := width - len(maxStr)
	if dotsCount < 0 {
		dotsCount = 0
	}
	if len(lines) > 0 {
		lines[0] = fmt.Sprintf(" [yellow]%s[green]", maxStr) + strings.Repeat("·", dotsCount) + "[-]"
	}

	return strings.Join(lines, "\n")
}

func (a *App) updateConnList() {
	defer func() {
		if r := recover(); r != nil {
			f, _ := os.OpenFile(globalErrorLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if f != nil {
				fmt.Fprintf(f, "=== updateConnList PANIC ===\n%v\n%s\n\n", r, debug.Stack())
				f.Close()
			}
		}
	}()

	a.connList.Clear()

	// Header
	headers := []string{"Protocol", "Destination", "Upload", "Download"}
	for i, h := range headers {
		cell := tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false)
		a.connList.SetCell(0, i, cell)
	}

	a.mu.RLock()
	if !a.isConnected {
		a.mu.RUnlock()
		return
	}
	stats := a.lastStats
	a.mu.RUnlock()

	if stats == nil {
		return
	}

	row := 1
	// Display inbound stats
	for tag, inbound := range stats.Inbounds {
		a.connList.SetCell(row, 0, tview.NewTableCell(tag).SetTextColor(tcell.ColorWhite))
		a.connList.SetCell(row, 1, tview.NewTableCell("-").SetTextColor(tcell.ColorGray))
		a.connList.SetCell(row, 2, tview.NewTableCell(formatBytes(inbound.Uplink)).SetTextColor(tcell.ColorGreen))
		a.connList.SetCell(row, 3, tview.NewTableCell(formatBytes(inbound.Downlink)).SetTextColor(tcell.ColorAqua))
		row++
	}

	// Display outbound stats
	for tag, outbound := range stats.Outbounds {
		a.connList.SetCell(row, 0, tview.NewTableCell(tag).SetTextColor(tcell.ColorWhite))
		a.connList.SetCell(row, 1, tview.NewTableCell("-").SetTextColor(tcell.ColorGray))
		a.connList.SetCell(row, 2, tview.NewTableCell(formatBytes(outbound.Uplink)).SetTextColor(tcell.ColorGreen))
		a.connList.SetCell(row, 3, tview.NewTableCell(formatBytes(outbound.Downlink)).SetTextColor(tcell.ColorAqua))
		row++
	}
}

func (a *App) updateServerList() {
	defer func() {
		if r := recover(); r != nil {
			f, _ := os.OpenFile(globalErrorLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if f != nil {
				fmt.Fprintf(f, "=== updateServerList PANIC ===\n%v\n%s\n\n", r, debug.Stack())
				f.Close()
			}
		}
	}()

	a.serverList.Clear()
	a.mu.RLock()
	servers := make([]*superray.Server, len(a.servers))
	copy(servers, a.servers)
	currentServer := a.currentServer
	a.mu.RUnlock()

	for i, s := range servers {
		name := s.Name
		if name == "" {
			name = fmt.Sprintf("%s:%d", maskAddress(s.Address), s.Port)
		}

		// Add latency indicator (escape brackets for tview)
		latencyStr := ""
		if s.Latency > 0 {
			latencyStr = fmt.Sprintf(" [[%dms[]]", s.Latency)
		} else if s.Latency == -1 {
			latencyStr = " [[timeout[]]"
		}

		text := fmt.Sprintf("[[%s[]] %s%s", strings.ToUpper(s.Protocol), name, latencyStr)
		a.serverList.AddItem(text, "", 0, nil)

		if currentServer != nil && s.Address == currentServer.Address && s.Port == currentServer.Port {
			a.serverList.SetCurrentItem(i)
		}
	}
}

func (a *App) connectSelected() {
	a.log("[green]Connect key pressed[white]")
	index := a.serverList.GetCurrentItem()

	if index < 0 {
		a.log("[yellow]No server selected. Press 's' to add subscription first.[white]")
		return
	}
	a.mu.RLock()
	serverCount := len(a.servers)
	a.mu.RUnlock()

	if serverCount == 0 {
		a.log("[yellow]No servers available. Press 's' to add subscription, then 'r' to refresh.[white]")
		return
	}
	idx := index
	safeGo(func() { a.connectToServer(idx) })
}

func (a *App) connectToServer(index int) {
	a.mu.Lock()
	if a.isConnecting {
		a.mu.Unlock()
		a.log("[yellow]Already connecting...[white]")
		return
	}
	a.isConnecting = true
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		a.isConnecting = false
		a.mu.Unlock()
	}()

	a.log(fmt.Sprintf("[green]Connecting to server %d...[white]", index))

	a.mu.RLock()
	serverCount := len(a.servers)
	if index >= serverCount {
		a.mu.RUnlock()
		a.log(fmt.Sprintf("[red]Invalid server index %d (total: %d)[white]", index, serverCount))
		return
	}
	server := a.servers[index]
	a.mu.RUnlock()

	// Disconnect existing connection
	a.disconnectSync()

	a.log(fmt.Sprintf("Connecting to %s (%s:%d) [%s]...", server.Name, maskAddress(server.Address), server.Port, server.Protocol))

	// Build config using SuperRay library
	config := a.buildConfig(server)

	// Start Xray
	type result struct {
		id  string
		err error
	}
	ch := make(chan result, 1)
	safeGo(func() {
		id, err := superray.Run(config)
		ch <- result{id, err}
	})

	select {
	case res := <-ch:
		if res.err != nil {
			a.log(fmt.Sprintf("[red]Failed to connect: %v[white]", res.err))
			return
		}

		a.mu.Lock()
		a.instanceID = res.id
		a.currentServer = server
		a.isConnected = true
		a.totalUpload = 0
		a.totalDownload = 0
		a.trafficHistory = make([]TrafficPoint, 0, a.historyMaxLen)
		a.uploadSpeed = 0
		a.downloadSpeed = 0
		a.lastStats = nil
		a.mu.Unlock()

		a.log(fmt.Sprintf("[green]Connected to %s[white]", server.Name))
		if a.tunMode {
			a.log("[aqua]Mode: TUN (global proxy)[white]")
			a.log(fmt.Sprintf("[darkgray]SOCKS5: 127.0.0.1:%d | HTTP: 127.0.0.1:%d[white]", a.localPort, a.localPort+1))
			// Set up TUN device connected to Xray instance
			safeGo(func() {
				a.startTUN(res.id)
			})
		} else {
			a.log(fmt.Sprintf("[darkgray]SOCKS5: 127.0.0.1:%d | HTTP: 127.0.0.1:%d[white]", a.localPort, a.localPort+1))
		}

		a.forceRefresh()

	case <-time.After(30 * time.Second):
		a.log("[red]Connection timeout[white]")
	}
}

func (a *App) disconnectSync() {
	a.mu.Lock()
	if !a.isConnected {
		a.mu.Unlock()
		return
	}

	instanceID := a.instanceID
	a.isConnected = false
	a.instanceID = ""
	a.currentServer = nil
	a.mu.Unlock()

	if instanceID != "" {
		superray.DestroyInstance(instanceID)
	}
}

func (a *App) disconnect() {
	a.mu.Lock()
	if !a.isConnected {
		a.mu.Unlock()
		return
	}

	instanceID := a.instanceID
	tunMode := a.tunMode
	a.isConnected = false
	a.instanceID = ""
	a.currentServer = nil
	a.lastStats = nil
	a.mu.Unlock()

	// Stop TUN and cleanup routes if in TUN mode
	if tunMode {
		a.stopTUN()
	}

	if instanceID != "" {
		superray.DestroyInstance(instanceID)
	}

	a.log("[yellow]Disconnected[white]")

	a.app.QueueUpdateDraw(func() {
		a.updateStatusView()
		a.updateSpeedView()
	})
}

func (a *App) buildConfig(server *superray.Server) string {
	// Build outbound from server
	outbound := buildOutboundFromServer(server, "proxy")

	// Build routing rules
	var rules []map[string]interface{}

	// 1. Private and reserved addresses direct
	rules = append(rules, map[string]interface{}{
		"type": "field",
		"ip": []string{
			"10.0.0.0/8",
			"172.16.0.0/12",
			"192.168.0.0/16",
			"127.0.0.0/8",
			"100.64.0.0/10",
			"169.254.0.0/16",
			"224.0.0.0/4",
			"240.0.0.0/4",
			"255.255.255.255/32",
			"::1/128",
			"fc00::/7",
			"fe80::/10",
		},
		"outboundTag": "direct",
	})

	// 2. Country-based direct
	if len(a.directCountries) > 0 {
		var geoips []string
		for _, country := range a.directCountries {
			geoips = append(geoips, "geoip:"+country)
		}
		rules = append(rules, map[string]interface{}{
			"type":        "field",
			"ip":          geoips,
			"outboundTag": "direct",
		})
	}

	// 3. Default: all other traffic through proxy
	rules = append(rules, map[string]interface{}{
		"type":        "field",
		"network":     "tcp,udp",
		"outboundTag": "proxy",
	})

	// Build inbounds based on mode
	var inbounds []map[string]interface{}

	if a.tunMode {
		// TUN mode: use TUN device for global proxy
		inbounds = append(inbounds, map[string]interface{}{
			"tag":      "tun-in",
			"protocol": "dokodemo-door",
			"listen":   "127.0.0.1",
			"port":     a.localPort + 10,
			"settings": map[string]interface{}{
				"network":        "tcp,udp",
				"followRedirect": true,
			},
			"sniffing": map[string]interface{}{
				"enabled":      true,
				"destOverride": []string{"http", "tls", "quic"},
			},
		})
	}

	// Always add SOCKS and HTTP inbounds
	inbounds = append(inbounds, map[string]interface{}{
		"tag":      "socks-in",
		"protocol": "socks",
		"listen":   "127.0.0.1",
		"port":     a.localPort,
		"settings": map[string]interface{}{
			"udp": true,
		},
	})
	inbounds = append(inbounds, map[string]interface{}{
		"tag":      "http-in",
		"protocol": "http",
		"listen":   "127.0.0.1",
		"port":     a.localPort + 1,
	})

	config := map[string]interface{}{
		"stats": map[string]interface{}{},
		"policy": map[string]interface{}{
			"system": map[string]interface{}{
				"statsInboundUplink":    true,
				"statsInboundDownlink":  true,
				"statsOutboundUplink":   true,
				"statsOutboundDownlink": true,
			},
		},
		"log": map[string]interface{}{
			"loglevel": "warning",
			"access":   a.accessLogPath,
			"error":    a.errorLogPath,
		},
		"inbounds": inbounds,
		"outbounds": []interface{}{
			outbound,
			map[string]interface{}{
				"tag":      "direct",
				"protocol": "freedom",
			},
			map[string]interface{}{
				"tag":      "block",
				"protocol": "blackhole",
			},
		},
		"routing": map[string]interface{}{
			"domainStrategy": "IPIfNonMatch",
			"rules":          rules,
		},
	}

	configJSON, _ := json.Marshal(config)
	return string(configJSON)
}

func (a *App) loadSubscription() {
	if a.subscriptionURL == "" {
		a.log("[yellow]No subscription URL set. Press 's' to add one.[white]")
		return
	}

	a.log("Loading subscription...")

	// Add and update subscription using SuperRay library
	if err := superray.AddSubscription("default", a.subscriptionURL); err != nil {
		// Might already exist, try update
	}

	servers, err := superray.UpdateSubscription("default")
	if err != nil {
		a.log(fmt.Sprintf("[red]Failed to update subscription: %v[white]", err))
		return
	}

	a.mu.Lock()
	a.servers = servers
	a.mu.Unlock()

	a.log(fmt.Sprintf("[green]Loaded %d servers[white]", len(servers)))

	a.app.QueueUpdateDraw(func() {
		a.updateServerList()
	})
}

func (a *App) testLatency() {
	a.mu.RLock()
	servers := a.servers
	a.mu.RUnlock()

	if len(servers) == 0 {
		a.log("[yellow]No servers to test[white]")
		return
	}

	a.log("Testing latency...")

	// Build server list for batch test
	serverList := make([]map[string]interface{}, len(servers))
	for i, s := range servers {
		serverList[i] = map[string]interface{}{
			"address": s.Address,
			"port":    s.Port,
			"name":    s.Name,
		}
	}

	results, err := superray.BatchLatencyTest(serverList, 10, 1, 5000)
	if err != nil {
		a.log(fmt.Sprintf("[red]Failed to test latency: %v[white]", err))
		return
	}

	// Update servers with latency
	a.mu.Lock()
	for i := range a.servers {
		for _, r := range results {
			if a.servers[i].Address == r.Address && a.servers[i].Port == r.Port {
				if r.Success {
					a.servers[i].Latency = int64(r.Latency)
				} else {
					a.servers[i].Latency = -1 // timeout
				}
				break
			}
		}
	}

	// Sort by latency (lower is better, timeout/untested at end)
	sort.Slice(a.servers, func(i, j int) bool {
		iLatency := a.servers[i].Latency
		jLatency := a.servers[j].Latency

		// Both timeout or not tested, keep original order
		if iLatency <= 0 && jLatency <= 0 {
			return false
		}
		// i is timeout/untested, put it after j
		if iLatency <= 0 {
			return false
		}
		// j is timeout/untested, put i before j
		if jLatency <= 0 {
			return true
		}
		// Compare actual latency values
		return iLatency < jLatency
	})
	a.mu.Unlock()

	a.log("[green]Latency test complete[white]")

	a.app.QueueUpdateDraw(func() {
		a.updateServerList()
	})
}

func (a *App) showSubscriptionDialog() {
	input := tview.NewInputField().
		SetLabel("Subscription URL: ").
		SetFieldWidth(60).
		SetText(a.subscriptionURL)

	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			a.subscriptionURL = input.GetText()
			a.pages.RemovePage("modal")
			safeGo(func() { a.loadSubscription() })
		} else if key == tcell.KeyEscape {
			a.pages.RemovePage("modal")
		}
	})

	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(input, 3, 0, true).
			AddItem(nil, 0, 1, false), 70, 0, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("modal", modal, true, true)
	a.app.SetFocus(input)
}

func (a *App) log(msg string) {
	timestamp := time.Now().Format("15:04:05")
	a.app.QueueUpdate(func() {
		fmt.Fprintf(a.logView, "[darkgray]%s[-] %s\n", timestamp, msg)
		a.logView.ScrollToEnd()
	})
}

func (a *App) toggleTunMode() {
	a.log("[green]Toggle TUN mode pressed[white]")

	a.mu.Lock()
	wasConnected := a.isConnected
	currentMode := a.tunMode
	a.mu.Unlock()

	if !currentMode {
		if os.Geteuid() != 0 {
			a.log("[red]TUN mode requires admin/root privileges[white]")
			a.log("[yellow]Please run with: sudo ./superray-tui[white]")
			return
		}
	}

	if wasConnected {
		a.log("[yellow]Disconnecting before switching mode...[white]")
		a.disconnect()
	}

	a.mu.Lock()
	a.tunMode = !a.tunMode
	mode := a.tunMode
	a.mu.Unlock()

	if mode {
		a.log("[green]Switched to TUN mode (global proxy)[white]")
		a.log("[yellow]Traffic will be routed through TUN device[white]")
	} else {
		a.log("[yellow]Switched to Proxy mode (SOCKS5/HTTP)[white]")
		// Clean up TUN device if it was created
		superray.CloseAllTUNDevices()
	}

	a.app.QueueUpdateDraw(func() {
		a.updateStatusView()
		a.updateSpeedView()
		a.updateConnList()
		if a.screen != nil {
			a.screen.Sync()
		}
	})
}

func (a *App) startTUN(instanceID string) {
	a.log("[green]Starting System TUN device...[white]")

	// Get current server address for routing
	a.mu.RLock()
	currentServer := a.currentServer
	a.mu.RUnlock()

	if currentServer == nil {
		a.log("[red]No server connected for TUN mode[white]")
		return
	}

	// Clean up any existing TUN device first
	superray.CloseAllSystemTUNs()

	a.log(fmt.Sprintf("[darkgray]TUN config: MTU=1500, Addr=10.255.0.1/24[white]"))

	// Step 1: Create system TUN device (requires root)
	tunInfo, err := superray.CreateSystemTUN("tun0", []string{"10.255.0.1/24"}, 1500)
	if err != nil {
		a.log(fmt.Sprintf("[red]TUN creation error: %v[white]", err))
		a.log("[yellow]TUN mode requires root privileges. Run with: sudo ./superray-tui[white]")
		return
	}

	a.log(fmt.Sprintf("[green]TUN device created: %s[white]", tunInfo.Name))

	// Step 2: Start TUN stack connected to Xray instance
	if err := superray.StartSystemTUNStack("tun0", instanceID, "proxy"); err != nil {
		a.log(fmt.Sprintf("[red]TUN stack error: %v[white]", err))
		superray.CloseSystemTUN("tun0")
		return
	}
	a.log("[green]TUN stack started - traffic forwarding active[white]")

	// Step 3: Setup system routes
	a.log("[yellow]Setting up routes...[white]")
	if err := superray.SetupRoutes("tun0", currentServer.Address); err != nil {
		a.log(fmt.Sprintf("[yellow]Route setup warning: %v[white]", err))
		a.log("[darkgray]TUN device works, but routes may need manual configuration[white]")
	} else {
		a.log("[green]Routes configured - all traffic now goes through VPN[white]")
	}
}

func (a *App) stopTUN() {
	a.log("[yellow]Stopping TUN device...[white]")

	// Cleanup routes first
	if err := superray.CleanupRoutes("tun0"); err != nil {
		// Ignore errors, routes might not exist
	}

	// Close system TUN device
	if err := superray.CloseSystemTUN("tun0"); err != nil {
		// Try closing all as fallback
		superray.CloseAllSystemTUNs()
	}

	// Also cleanup any callback TUNs that might exist
	superray.CloseAllCallbackTUNs()
	superray.CloseAllTUNDevices()

	a.log("[yellow]TUN device closed, routes restored[white]")
}

func (a *App) showVersion() {
	version, _ := superray.Version()
	xrayVersion, _ := superray.XrayVersion()
	a.log(fmt.Sprintf("[cyan]SuperRay %s (Xray-core %s)[white]", version, xrayVersion))
}

// Helper functions

func formatSpeed(bytesPerSec float64) string {
	if bytesPerSec < 1024 {
		return fmt.Sprintf("%.0f B/s", bytesPerSec)
	} else if bytesPerSec < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", bytesPerSec/1024)
	} else if bytesPerSec < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB/s", bytesPerSec/1024/1024)
	}
	return fmt.Sprintf("%.2f GB/s", bytesPerSec/1024/1024/1024)
}

func formatBytes(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(bytes)/1024/1024)
	}
	return fmt.Sprintf("%.2f GB", float64(bytes)/1024/1024/1024)
}

// buildOutboundFromServer builds an outbound config from Server
func buildOutboundFromServer(server *superray.Server, tag string) map[string]interface{} {
	if tag == "" {
		tag = "proxy"
	}

	switch server.Protocol {
	case "vmess":
		security := server.Security
		if security == "" {
			security = "auto"
		}
		return map[string]interface{}{
			"protocol": "vmess",
			"tag":      tag,
			"settings": map[string]interface{}{
				"vnext": []map[string]interface{}{
					{
						"address": server.Address,
						"port":    server.Port,
						"users": []map[string]interface{}{
							{
								"id":       server.UUID,
								"alterId":  0,
								"security": security,
							},
						},
					},
				},
			},
			"streamSettings": buildStreamSettings(server),
		}

	case "vless":
		return map[string]interface{}{
			"protocol": "vless",
			"tag":      tag,
			"settings": map[string]interface{}{
				"vnext": []map[string]interface{}{
					{
						"address": server.Address,
						"port":    server.Port,
						"users": []map[string]interface{}{
							{
								"id":         server.UUID,
								"encryption": "none",
								"flow":       server.Flow,
							},
						},
					},
				},
			},
			"streamSettings": buildStreamSettings(server),
		}

	case "trojan":
		return map[string]interface{}{
			"protocol": "trojan",
			"tag":      tag,
			"settings": map[string]interface{}{
				"servers": []map[string]interface{}{
					{
						"address":  server.Address,
						"port":     server.Port,
						"password": server.Password,
					},
				},
			},
			"streamSettings": buildStreamSettings(server),
		}

	case "shadowsocks", "ss":
		return map[string]interface{}{
			"protocol": "shadowsocks",
			"tag":      tag,
			"settings": map[string]interface{}{
				"servers": []map[string]interface{}{
					{
						"address":  server.Address,
						"port":     server.Port,
						"method":   server.Method,
						"password": server.Password,
					},
				},
			},
		}
	}

	return nil
}

// buildStreamSettings builds stream settings from Server
func buildStreamSettings(server *superray.Server) map[string]interface{} {
	stream := map[string]interface{}{
		"network": server.Network,
	}

	if server.Network == "" {
		stream["network"] = "tcp"
	}

	// TLS settings
	if server.TLS == "tls" || server.TLS == "true" {
		stream["security"] = "tls"
		tlsSettings := map[string]interface{}{}
		if server.SNI != "" {
			tlsSettings["serverName"] = server.SNI
		}
		if server.Fingerprint != "" {
			tlsSettings["fingerprint"] = server.Fingerprint
		}
		stream["tlsSettings"] = tlsSettings
	} else if server.Security == "reality" {
		stream["security"] = "reality"
		stream["realitySettings"] = map[string]interface{}{
			"serverName":  server.SNI,
			"fingerprint": server.Fingerprint,
			"publicKey":   server.PublicKey,
			"shortId":     server.ShortID,
		}
	}

	// Network-specific settings
	switch server.Network {
	case "ws":
		wsSettings := map[string]interface{}{}
		if server.Path != "" {
			wsSettings["path"] = server.Path
		}
		if server.Host != "" {
			wsSettings["headers"] = map[string]interface{}{
				"Host": server.Host,
			}
		}
		stream["wsSettings"] = wsSettings

	case "grpc":
		grpcSettings := map[string]interface{}{}
		if server.Path != "" {
			grpcSettings["serviceName"] = server.Path
		}
		stream["grpcSettings"] = grpcSettings

	case "h2", "http":
		httpSettings := map[string]interface{}{}
		if server.Path != "" {
			httpSettings["path"] = server.Path
		}
		if server.Host != "" {
			httpSettings["host"] = []string{server.Host}
		}
		stream["httpSettings"] = httpSettings
	}

	return stream
}

// lookupServerGeoIP looks up GeoIP info for a server address
func (a *App) lookupServerGeoIP(addr string) {
	info := lookupIP(addr)
	a.mu.Lock()
	a.serverGeoInfo = info
	a.mu.Unlock()

	// Update UI
	a.app.QueueUpdateDraw(func() {
		a.updateStatusView()
	})
}

// lookupIP looks up an IP address using ip-api.com and returns detailed geo information
func lookupIP(ipStr string) *GeoIPInfo {
	// Check if it's a hostname, resolve it first
	ip := net.ParseIP(ipStr)
	if ip == nil {
		// Try to resolve hostname
		ips, err := net.LookupIP(ipStr)
		if err != nil || len(ips) == 0 {
			return &GeoIPInfo{
				IP:      ipStr,
				Country: "Unknown",
				Status:  "fail",
			}
		}
		ipStr = ips[0].String()
	}

	// Check cache
	ipCacheMu.RLock()
	if cached, ok := ipCache[ipStr]; ok {
		if time.Since(ipCacheTime[ipStr]) < cacheTTL {
			ipCacheMu.RUnlock()
			return cached
		}
	}
	ipCacheMu.RUnlock()

	// Query ip-api.com (free, no API key needed, 45 requests/minute limit)
	info := queryIPAPI(ipStr)

	// Cache result
	ipCacheMu.Lock()
	ipCache[ipStr] = info
	ipCacheTime[ipStr] = time.Now()
	ipCacheMu.Unlock()

	return info
}

// queryIPAPI queries ip-api.com for IP information
func queryIPAPI(ipStr string) *GeoIPInfo {
	info := &GeoIPInfo{
		IP:      ipStr,
		Country: "Unknown",
		Status:  "fail",
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Use fields parameter to get specific info
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,countryCode,region,regionName,city,isp,org,as,asname,query", ipStr)

	resp, err := client.Get(url)
	if err != nil {
		return info
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return info
	}

	if err := json.NewDecoder(resp.Body).Decode(info); err != nil {
		return info
	}

	return info
}

// FormatGeoInfo formats GeoIPInfo for display
func (g *GeoIPInfo) FormatGeoInfo() string {
	if g.Status != "success" {
		return "Unknown"
	}

	result := ""
	if g.Country != "" {
		result = g.Country
	}
	if g.City != "" && g.City != g.Country {
		result += " " + g.City
	}
	return result
}

// FormatOrg formats organization/ISP info for display
func (g *GeoIPInfo) FormatOrg() string {
	if g.Org != "" {
		return g.Org
	}
	if g.ISP != "" {
		return g.ISP
	}
	if g.ASName != "" {
		return g.ASName
	}
	return ""
}
