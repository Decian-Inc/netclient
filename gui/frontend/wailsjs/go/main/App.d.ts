// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {main} from '../models';
import {config} from '../models';
import {wgtypes} from '../models';
import {frontend} from '../models';

export function GoConnectToNetwork(arg1:string):Promise<any>;

export function GoDisconnectFromNetwork(arg1:string):Promise<any>;

export function GoGetKnownNetworks():Promise<Array<main.Network>>;

export function GoGetNetclientConfig():Promise<main.NcConfig>;

export function GoGetNetwork(arg1:string):Promise<main.Network>;

export function GoGetNodePeers(arg1:config.Node):Promise<Array<wgtypes.PeerConfig>>;

export function GoGetRecentServerNames():Promise<Array<string>>;

export function GoGetStatus():Promise<any>;

export function GoJoinNetworkByBasicAuth(arg1:string,arg2:string,arg3:string,arg4:string):Promise<any>;

export function GoJoinNetworkBySso(arg1:string,arg2:string):Promise<main.SsoJoinResDto>;

export function GoLeaveNetwork(arg1:string):Promise<any>;

export function GoOpenDialogue(arg1:frontend.DialogType,arg2:string,arg3:string):Promise<string>;

export function GoPullLatestNodeConfig(arg1:string):Promise<main.Network>;

export function GoRegisterWithEnrollmentKey(arg1:string):Promise<any>;

export function GoUninstall():Promise<any>;

export function GoUpdateNetclientConfig(arg1:config.Config):Promise<any>;

export function GoWriteToClipboard(arg1:string):Promise<any>;
