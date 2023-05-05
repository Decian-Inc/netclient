package config

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"sync"

	"github.com/gravitl/netclient/nmproxy/common"
	proxyModels "github.com/gravitl/netclient/nmproxy/models"
	"github.com/gravitl/netmaker/logger"
	"github.com/gravitl/netmaker/models"
)

var (
	// contains all the config related to proxy
	config        = &Config{}
	natAutoSwitch bool
	// DumpSignalChan - channel to signal dump proxy conns info
	DumpSignalChan = make(chan struct{}, 5)
)

// Config - struct for proxy config
type Config struct {
	HostInfo                proxyModels.HostInfo
	ProxyStatus             bool
	mutex                   *sync.RWMutex
	ifaceConfig             wgIfaceConf
	settings                map[string]proxyModels.Settings // host settings per server
	metricsThreadDone       context.CancelFunc
	metricsCollectionStatus bool
	serverConn              *net.UDPConn
	fireWallStatus          bool
	fireWallClose           func()
}
type proxyPeerConn struct {
	PeerPublicKey       string `json:"peer_public_key"`
	PeerEndpoint        string `json:"peer_endpoint"`
	ProxyEndpoint       string `json:"proxy_endpoint"`
	ProxyRemoteEndpoint string `json:"proxy_remote_endpoint"`
}

// InitializeCfg - intializes all the variables and sets defaults
func InitializeCfg() {
	config = &Config{
		ProxyStatus: true,
		mutex:       &sync.RWMutex{},
		ifaceConfig: wgIfaceConf{
			iface:        nil,
			turnPeerMap:  make(map[string]map[string]proxyModels.TurnPeerCfg),
			hostTurnCfg:  make(map[string]proxyModels.TurnCfg),
			proxyPeerMap: make(proxyModels.PeerConnMap),
			peerHashMap:  make(map[string]*proxyModels.RemotePeer),
			relayPeerMap: make(map[string]map[string]*proxyModels.RemotePeer),
			allPeersConf: make(map[string]models.HostPeerMap),
		},
		settings: make(map[string]proxyModels.Settings),
	}
}

// Config.IsProxyRunning - checks if proxy is running
func (c *Config) IsProxyRunning() bool {
	return c.ProxyStatus
}

// Config.SetHostInfo - sets host info
func (c *Config) SetHostInfo(hostInfo proxyModels.HostInfo) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.HostInfo = hostInfo
}

// Config.StopMetricsCollectionThread - stops the metrics thread // only when host proxy is disabled
func (c *Config) StopMetricsCollectionThread() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.metricsThreadDone != nil {
		c.metricsThreadDone()
		c.metricsCollectionStatus = false
	}
}

// Config.GetMetricsCollectionStatus - fetchs metrics collection status when proxy is disabled for host
func (c *Config) GetMetricsCollectionStatus() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.metricsCollectionStatus
}

// Config.SetMetricsThreadCtx - sets the metrics thread ctx
func (c *Config) SetMetricsThreadCtx(cancelFunc context.CancelFunc) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.metricsThreadDone = cancelFunc
	c.metricsCollectionStatus = true
}

// Config.GetHostInfo - gets the host info
func (c *Config) GetHostInfo() proxyModels.HostInfo {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.HostInfo
}

// Reset - resets Config // to be called only when proxy is shutting down
func Reset() {
	config = &Config{}
}

// GetCfg - fethes Config
func GetCfg() *Config {
	return config
}

// Config.GetSettings - fetches host settings
func (c *Config) GetSettings(server string) proxyModels.Settings {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if settings, ok := c.settings[server]; ok {
		return settings
	}
	return proxyModels.Settings{}
}

// Config.UpdateSettings - updates network settings
func (c *Config) UpdateSettings(server string, settings proxyModels.Settings) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.settings[server] = settings
}

// Config.SetRelayStatus - sets host relay status
func (c *Config) SetRelayStatus(server string, value bool) {
	settings := c.GetSettings(server)
	settings.IsRelay = value
	c.UpdateSettings(server, settings)
}

// Config.IsRelay - fetches relay status value of the host
func (c *Config) IsRelay(server string) bool {
	return c.GetSettings(server).IsRelay
}

// Config.IsGlobalRelay - checks if host relay globally
func (c *Config) IsGlobalRelay() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	for _, settings := range c.settings {
		if settings.IsRelay {
			return true
		}
	}
	return false
}

// Config.SetIngressGwStatus - sets ingressGW status
func (c *Config) SetIngressGwStatus(server string, value bool) {
	settings := c.GetSettings(server)
	settings.IsIngressGateway = value
	c.UpdateSettings(server, settings)
}

// Config.IsIngressGw - checks if ingressGW by server
func (c *Config) IsIngressGw(server string) bool {

	return c.GetSettings(server).IsIngressGateway
}

// Config.IsEgressGw - checks if egressGW by server
func (c *Config) IsEgressGw(server string) bool {
	return c.GetSettings(server).IsEgressGateway
}

// Config.SetEgressGwStatus - sets egressGW status
func (c *Config) SetEgressGwStatus(server string, value bool) {
	settings := c.GetSettings(server)
	settings.IsEgressGateway = value
	c.UpdateSettings(server, settings)
}

// Config.SetRelayedStatus - sets relayed status
func (c *Config) SetRelayedStatus(server string, value bool) {
	settings := c.GetSettings(server)
	settings.IsRelayed = value
	c.UpdateSettings(server, settings)
}

// Config.GetRelayedStatus - gets relayed status
func (c *Config) GetRelayedStatus(server string) bool {
	return c.GetSettings(server).IsRelayed
}

// NatAutoSwitchDone - check if nat automatically switched on already for devices behind NAT
func NatAutoSwitchDone() bool {
	return natAutoSwitch
}

// SetNatAutoSwitch - set NAT auto switch to true
func SetNatAutoSwitch() {
	natAutoSwitch = true
}

// Config.ShouldUseProxy - checks if proxy is running behind NAT
func (c *Config) ShouldUseProxy() bool {
	return c.HostInfo.NatType == models.NAT_Types.Asymmetric || c.HostInfo.NatType == models.NAT_Types.Double
}

// Config.GetServerConn - fetches the server connection
func (c *Config) GetServerConn() *net.UDPConn {
	return c.serverConn
}

// Config.SetServerConn - sets server connection
func (c *Config) SetServerConn(conn *net.UDPConn) {
	c.serverConn = conn
}

// Config.SetFwStatus - sets the firewall status
func (c *Config) SetFwStatus(s bool) {
	c.fireWallStatus = s
}

// Config.SetFwCloseFunc - sets the firewall flush func
func (c *Config) SetFwCloseFunc(fwFlush func()) {
	c.fireWallClose = fwFlush
}

// Config.GetFwStatus - gets the firewall status
func (c *Config) GetFwStatus() bool {
	return c.fireWallStatus
}

// Config.StopFw - flushes all the firewall rules
func (c *Config) StopFw() {
	c.fireWallClose()
}

// Config.Dump - dumps the proxy peer connections information
func (c *Config) Dump() {
	peersConn := c.GetAllProxyPeers()
	proxyConns := []proxyPeerConn{}
	for peerPubKey, peerI := range peersConn {
		peerConnI := proxyPeerConn{
			PeerPublicKey: peerPubKey,
		}

		if peerI.Config.PeerConf.Endpoint != nil {
			peerConnI.PeerEndpoint = peerI.Config.PeerConf.Endpoint.String()
		}
		if peerI.Config.LocalConnAddr != nil {
			peerConnI.ProxyEndpoint = peerI.Config.LocalConnAddr.String()
		}
		if peerI.Config.RemoteConnAddr != nil {
			peerConnI.ProxyRemoteEndpoint = peerI.Config.RemoteConnAddr.String()
		}
		proxyConns = append(proxyConns, peerConnI)
	}
	out, err := json.MarshalIndent(proxyConns, "", " ")
	if err != nil {
		logger.Log(0, "failed to marshal list output: ", err.Error())
	}
	os.WriteFile(common.GetDataPath()+"proxy.json", out, os.ModePerm)
}
