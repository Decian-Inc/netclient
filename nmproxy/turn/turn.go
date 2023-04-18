package turn

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/devilcove/httpclient"
	"github.com/gravitl/netclient/auth"
	ncconfig "github.com/gravitl/netclient/config"
	"github.com/gravitl/netclient/nmproxy/config"
	"github.com/gravitl/netclient/nmproxy/models"
	"github.com/gravitl/netmaker/logger"
	"github.com/gravitl/netmaker/logic"
	nm_models "github.com/gravitl/netmaker/models"
	"github.com/pion/logging"
	"github.com/pion/turn"
)

// StartClient - starts the turn client on the netclient for the peer
func StartClient(turnDomain string, turnPort int) (net.PacketConn, error) {
	conn, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		logger.Log(0, "Failed to listen: %s", err.Error())
		return nil, err
	}
	turnServerAddr := fmt.Sprintf("%s:%d", turnDomain, turnPort)
	cfg := &turn.ClientConfig{
		STUNServerAddr: turnServerAddr,
		TURNServerAddr: turnServerAddr,
		Conn:           conn,
		Username:       ncconfig.Netclient().ID.String(),
		Password:       logic.ConvHostPassToHash(ncconfig.Netclient().HostPass),
		Realm:          turnDomain,
		Software:       "netmaker",
		LoggerFactory:  logging.NewDefaultLoggerFactory(),
	}
	client, err := turn.NewClient(cfg)
	if err != nil {
		logger.Log(0, "Failed to create TURN client: %s", err.Error())
		conn.Close()
		return nil, err
	}
	err = client.Listen()
	if err != nil {
		logger.Log(0, "Failed to listen: %s", err.Error())
		conn.Close()
		client.Close()
		return nil, err
	}
	// allocate turn relay address to host for the peer and exchange information with peer
	turnConn, err := AllocateAddr(client)
	if err != nil {
		logger.Log(0, "failed to allocate addr on turn: ", err.Error())
		return nil, err
	}
	config.GetCfg().SetTurnCfg(models.TurnCfg{
		Cfg:      cfg,
		Client:   client,
		TurnConn: turnConn,
	})
	return turnConn, nil
}

func AllocateAddr(client *turn.Client) (net.PacketConn, error) {
	// Allocate a relay socket on the TURN server. On success, it
	// will return a net.PacketConn which represents the remote
	// socket.
	relayConn, err := client.Allocate()
	if err != nil {
		logger.Log(0, "Failed to allocate: ", err.Error())
		return nil, err
	}
	// Send BindingRequest to learn our external IP
	mappedAddr, err := client.SendBindingRequest()
	if err != nil {
		logger.Log(0, "failed to send binding req: ", err.Error())
		return nil, err
	}
	// Punch a UDP hole for the relayConn by sending a data to the mappedAddr.
	// This will trigger a TURN client to generate a permission request to the
	// TURN server. After this, packets from the IP address will be accepted by
	// the TURN server.
	_, err = relayConn.WriteTo([]byte("Hello"), mappedAddr)
	if err != nil {
		logger.Log(0, "failed to send binding request: ", err.Error())
		return nil, err
	}
	// The relayConn's local address is actually the transport
	// address assigned on the TURN server.
	log.Printf("relayed-address=%s", relayConn.LocalAddr().String())
	return relayConn, nil
}

func SignalPeer(serverName string, signal nm_models.Signal) error {
	server := ncconfig.GetServer(serverName)
	host := ncconfig.Netclient()
	if host == nil {
		return fmt.Errorf("no configured host found")
	}
	token, err := auth.Authenticate(server, host)
	if err != nil {
		return err
	}
	endpoint := httpclient.JSONEndpoint[nm_models.Signal, nm_models.ErrorResponse]{
		URL:           "https://" + server.API,
		Route:         fmt.Sprintf("/api/v1/host/%s/signalpeer", ncconfig.Netclient().ID.String()),
		Method:        http.MethodPost,
		Authorization: "Bearer " + token,
		Data:          signal,
		Response:      nm_models.Signal{},
		ErrorResponse: nm_models.ErrorResponse{},
	}
	_, errData, err := endpoint.GetJSON(nm_models.Signal{}, nm_models.ErrorResponse{})
	if err != nil {
		//if errors.Is(err, httpclient.ErrStatus) {
		logger.Log(0, "error signalling peer", strconv.Itoa(errData.Code), errData.Message)
		//}
		return err
	}
	return nil
}