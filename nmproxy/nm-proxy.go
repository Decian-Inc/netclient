package nmproxy

import (
	"context"
	"fmt"
	"sync"
	"time"

	ncconfig "github.com/gravitl/netclient/config"
	"github.com/gravitl/netclient/nmproxy/config"
	"github.com/gravitl/netclient/nmproxy/manager"
	ncmodels "github.com/gravitl/netclient/nmproxy/models"
	"github.com/gravitl/netclient/nmproxy/server"
	"github.com/gravitl/netclient/nmproxy/turn"
	"github.com/gravitl/netmaker/logger"
	"github.com/gravitl/netmaker/models"
)

// Start - setups the global cfg for proxy and starts the proxy server
func Start(ctx context.Context, wg *sync.WaitGroup,
	mgmChan chan *models.HostPeerUpdate, hostNatInfo *ncmodels.HostInfo, proxyPort int) {

	if config.GetCfg().IsProxyRunning() {
		logger.Log(1, "Proxy is running already...")
		return
	}
	defer wg.Done()
	if hostNatInfo == nil {
		return
	}
	if hostNatInfo.PrivIp == nil || hostNatInfo.PublicIp == nil {
		logger.Log(0, "failed to create proxy, check if stun list is configured correctly on your server")
		return
	}
	logger.Log(0, "Starting Proxy...")
	if proxyPort == 0 {
		proxyPort = models.NmProxyPort
	}
	config.InitializeCfg()
	defer config.Reset()
	logger.Log(0, fmt.Sprintf("set nat info: %v", hostNatInfo))
	config.GetCfg().SetHostInfo(*hostNatInfo)

	// start the netclient proxy server
	err := server.NmProxyServer.CreateProxyServer(proxyPort, 0, config.GetCfg().GetHostInfo().PrivIp.String())
	if err != nil {
		logger.FatalLog("failed to create proxy: ", err.Error())
	}
	config.GetCfg().SetServerConn(server.NmProxyServer.Server)
	wg.Add(1)
	go manager.Start(ctx, wg, mgmChan)
	wg.Add(1)
	go turn.WatchPeerSignals(ctx, wg)
	if turn.ShouldUseTurn(hostNatInfo.NatType) {
		time.Sleep(time.Second * 2) // add a delay for clients to send turn register message to server
		turn.Init(ctx, wg, ncconfig.GetAllTurnConfigs())
	}
	server.NmProxyServer.Listen(ctx)
}
